package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/fsnotify.v0"
	"gopkg.in/yaml.v1"
)

// Game contains the entire game.
type Game struct {
	name               string
	window             Window
	data               Data
	watcher            *fsnotify.Watcher
	touched            chan string
	running            bool
	handlingFileEvents bool
}

// Data stores all the stuff pulled in from YAML
type Data struct {
	Manifest Manifest
	Settings Settings
}

// Manifest contains global game properties.
type Manifest struct {
	Name string
}

// Settings holds options for the game.
type Settings struct {
	Width      int
	Height     int
	WindowMode windowMode
}

// windowMode specifies Fullscreen, Windowed, etc... for the Window.
type windowMode int

const (
	windowModeUnknown = iota - 1
	windowModeWindowed
	windowModeFullscreen
)

const (
	windowed   = "windowed"
	fullscreen = "fullscreen"
	unknown    = "unknown"
)

func (wm windowMode) String() string {
	switch wm {
	case windowModeWindowed:
		return windowed
	case windowModeFullscreen:
		return fullscreen
	default:
		return unknown
	}
}

func (wm *windowMode) SetYAML(tag string, value interface{}) (ok bool) {
	if t, ok := value.(string); ok {
		*wm = stringToWindowMode(t)
	}

	return true
}

func (wm *windowMode) GetYAML() (tag string, value interface{}) {
	return "", string(*wm)
}

func stringToWindowMode(mode string) windowMode {
	switch mode {
	case windowed:
		return windowModeWindowed
	case fullscreen:
		return windowModeFullscreen
	default:
		return windowModeUnknown
	}
}

// Failsafe name for the game befoure it's loaded from the data files
const failsafeGameName = "Unnamed"

// NewGame constructs a Game.
func NewGame() *Game {
	game := new(Game)
	game.running = true
	game.handlingFileEvents = false
	game.touched = make(chan string)
	game.watcher, _ = spawnWatcher()
	game.data = failsafeData()
	game.window = nil

	return game
}

func (game *Game) SetWindow(window Window) {
	if game.window != nil {
		game.window.Destroy()
	}

	game.window = window
	game.updateWindowSettings()
}

// SetDataDirectory adds all files in this directory to game's internal data and watches for changes
func (game *Game) SetDataDirectory(path string) {
	game.watcher.Add(path)
	go injectInitialFiles(game.touched, path)
	go game.forwardWatcherFileEvents()
}

// ApplyDataChanges merges in new changes from data.
func (game *Game) ApplyDataChanges(data *Data) {
	game.applyManifestChanges(data)
	game.applySettingsChanges(data)
	game.updateWindowSettings()
}

func (game *Game) applyManifestChanges(data *Data) {
	if data.Manifest.Name != failsafeGameName {
		game.data.Manifest.Name = data.Manifest.Name
	}
}

func (game *Game) applySettingsChanges(data *Data) {
	replaceIfPositive(&game.data.Settings.Width, &data.Settings.Width)
	replaceIfPositive(&game.data.Settings.Height, &data.Settings.Height)
	if game.data.Settings.WindowMode != windowModeUnknown {
		game.data.Settings.WindowMode = data.Settings.WindowMode
	}
}

// replaceIfPositive overrides oldValue with the value in newValue if newValue is positive
func replaceIfPositive(oldValue *int, newValue *int) {
	if *oldValue >= 0 {
		*oldValue = *newValue
	}
}

func (game *Game) updateWindowSettings() {
	if game.window == nil {
		return
	}

	game.window.SetTitle(game.data.Manifest.Name)
	game.window.SetSize(game.data.Settings.Width, game.data.Settings.Height)
}

func failsafeData() Data {
	return Data{
		Manifest{failsafeGameName},
		Settings{800, 600, windowModeWindowed},
	}
}

func magicData() Data {
	return Data{
		Manifest{failsafeGameName},
		Settings{-1, -1, windowModeUnknown},
	}
}

// Finish cleans up. Closes the games file watcher.
func (game *Game) Finish() {
	game.watcher.Close()
	if game.window != nil {
		game.window.Destroy()
	}
}

func spawnWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Warning: Could not create fsnotify watcher.")
	}

	return watcher, nil
}

func watchDirectory(watcher *fsnotify.Watcher, root string) {
}

func injectInitialFiles(touched chan string, root string) {
	filepath.Walk(root,
		func(path string, _ os.FileInfo, _ error) error {
			injectInitialFile(touched, path)
			return nil
		},
	)
}

func injectInitialFile(touched chan string, filename string) {
	if extensionIsYaml(filename) {
		touched <- filename
	}
}

func extensionIsYaml(filename string) bool {
	return path.Ext(filename) == ".yaml" || path.Ext(filename) == ".yml"
}

// forwardWatcherFileEvents recieves events from the file watcher and fowards them to the game's touched channel
func (game *Game) forwardWatcherFileEvents() {
	if game.handlingFileEvents == true {
		log.Println("ERROR: Already handling events.")
		return
	}

	if game.watcher == nil {
		log.Println("ERROR: Tried to recieve file events without a watcher.")
		return
	}

	game.handlingFileEvents = true

	for {
		select {
		case event := <-game.watcher.Events:
			// [TODO]: Use modification timestamps...
			// [TODO]: Deal with Vim. CHMOD, RENAME, REMOVE's them rather than WRITE...
			// [TODO]: Deal with spammy editors. Wait for a period of time befoure consuming while merging dupes...
			if event.Op&fsnotify.Remove != fsnotify.Remove {
				game.touched <- event.Name
			}
		case err := <-game.watcher.Errors:
			log.Println("ERROR: fsnotify watcher error:", err)
		}
	}
}

func (game *Game) consumeAllFileEvents() {
	for game.consumeFileEvent() {
		// No-op
	}
}

//consumeFileEvent consumes a single file event, returns false if where was none
func (game *Game) consumeFileEvent() bool {
	select {
	case filename := <-game.touched:
		if extensionIsYaml(filename) {
			data := magicData()
			data.loadYAML(filename)
			game.ApplyDataChanges(&data)
			game.watcher.Add(filename)
		}
		return true
	default:
		return false
	}
}

func dataDirectory() string {
	return path.Join("data")
}

// LoadYAML loads the data form a YAML file.
func (data *Data) loadYAML(filename string) error {
	log.Println("Loading YAML:", filename)
	rawdata, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = data.parseYaml(rawdata)
	return err
}

func (data *Data) parseYaml(raw []byte) error {
	err := yaml.Unmarshal(raw, &data)
	if err != nil {
		log.Println("Warning: Could not parse YAML.", err)
	}

	return err
}

// Run begins the game. Shows the Window, enters the main loop, etc...
func (game *Game) Run() {
	for game.running {
		game.consumeAllFileEvents()
		game.updateWindowSettings()
		if game.window != nil {
			game.window.Update()
		}
	}

	game.Finish()
}

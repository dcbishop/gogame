package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/fsnotify.v1"
	"gopkg.in/yaml.v1"
)

// Game contains the entire game.
type Game struct {
	name               string
	window             Window
	data               Data
	watcher            *fsnotify.Watcher
	handlingFileEvents bool
	quit               chan bool
	waitingFiles       []string
	Stdout             io.Writer
	Stderr             io.Writer
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
	windowModeUnknown    windowMode = iota - 1
	windowModeWindowed   windowMode = iota
	windowModeFullscreen windowMode = iota
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
	game.handlingFileEvents = false
	watcher, err := spawnWatcher()
	if err != nil {
		game.LogError("Could not create fsnotify watcher.")
	}
	game.watcher = watcher
	game.data = failsafeData()
	game.window = nil
	game.quit = make(chan bool)
	game.waitingFiles = []string{}
	game.Stdout = os.Stdout
	game.Stderr = os.Stderr

	return game
}

// SetWindow sets the window to render to.
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
	game.injectInitialFiles(path)
	game.forwardWatcherFileEvents()
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

// replaceIfPositive overrides oldValue with the value in newValue, if newValue is positive.
func replaceIfPositive(oldValue *int, newValue *int) {
	if *newValue >= 0 {
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

// Log will output an error to the game's StdOut writer.
func (game *Game) Log(a ...interface{}) {
	fmt.Fprintln(game.Stdout, a)
}

// LogError will output an error to the game's StdErr writer.
func (game *Game) LogError(a ...interface{}) {
	fmt.Fprintln(game.Stderr, a)
}

func spawnWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	return watcher, err
}

func (game *Game) injectInitialFiles(root string) {
	filepath.Walk(root,
		func(path string, _ os.FileInfo, _ error) error {
			game.injectInitialFile(path)
			return nil
		},
	)
}

func (game *Game) injectInitialFile(filename string) {
	if extensionIsYaml(filename) {
		game.waitingFiles = append(game.waitingFiles, filename)
	}
}

func extensionIsYaml(filename string) bool {
	return path.Ext(filename) == ".yaml" || path.Ext(filename) == ".yml"
}

// forwardWatcherFileEvents recieves events from the file watcher and fowards them to the game's touched channel
func (game *Game) forwardWatcherFileEvents() {
	if game.watcher == nil {
		game.LogError("Tried to recieve file events without a watcher.")
		return
	}

	events := true
	for events {
		select {
		case event := <-game.watcher.Events:
			// [TODO]: Use modification timestamps...
			// [TODO]: Deal with Vim. CHMOD, RENAME, REMOVE's them rather than WRITE...
			// [TODO]: Deal with spammy editors. Wait for a period of time befoure consuming while merging dupes...
			if event.Op&fsnotify.Remove != fsnotify.Remove {
				game.waitingFiles = append(game.waitingFiles, event.Name)
			}
		case err := <-game.watcher.Errors:
			game.LogError("fsnotify watcher error:", err)
		default:
			events = false
		}
	}
}

func (game *Game) consumeAllFileEvents() {
	for game.consumeFileEvent() {
		// No-op
	}
}

// consumeFileEvent consumes a single file event, returns false if where was none
func (game *Game) consumeFileEvent() bool {
	if len(game.waitingFiles) == 0 {
		return false
	}

	filename := game.popWaitingFile()
	game.processFile(filename)
	return true
}

func (game *Game) popWaitingFile() string {
	filename := game.waitingFiles[len(game.waitingFiles)-1]
	game.waitingFiles = game.waitingFiles[:len(game.waitingFiles)-1]
	return filename
}

func (game *Game) processFile(filename string) {
	if extensionIsYaml(filename) {
		data := magicData()
		game.Log("Loading:", filename, "...")
		err := data.loadYAML(filename)
		if err != nil {
			game.LogError("Could now load YAML:", err)
		} else {
			game.ApplyDataChanges(&data)
		}
		game.watcher.Add(filename)
	}
}

// DataDirectory returns the directory where the game should look for it's data.
func DataDirectory() string {
	return "data"
}

// LoadYAML loads the data form a YAML file.
func (data *Data) loadYAML(filename string) error {
	rawdata, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = data.parseYaml(rawdata)
	return err
}

func (data *Data) parseYaml(raw []byte) error {
	return yaml.Unmarshal(raw, &data)
}

// Run begins the game. Shows the Window, enters the main loop, etc...
func (game *Game) Run() {
	running := true

	for running {
		select {
		case _, _ = <-game.quit:
			running = false
		default:
			game.everyLoop()
		}
	}

	game.Finish()
}

func (game *Game) everyLoop() {
	game.forwardWatcherFileEvents()
	game.consumeAllFileEvents()
	game.updateWindowSettings()
	if game.window != nil {
		game.window.Update()
	}
}

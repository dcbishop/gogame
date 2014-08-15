package main

import (
	"errors"
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
	name    string
	window  Window
	data    Data
	watcher *fsnotify.Watcher
	touched chan string
	running bool
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
	Fullscreen bool
}

const GAMENAME_UNNAMED = "ERROR: No game name found."
const GAMENAME_LOADING = "Loading..."

// NewGame constructs a Game.
func NewGame() (*Game, error) {
	game := new(Game)
	game.running = true
	game.touched = make(chan string)
	game.data = failsafeData()

	go injectInitialFiles(game.touched, dataDirectory())
	game.consumeAllFileEvents()

	if game.data.Manifest.Name == GAMENAME_LOADING {
		game.data.Manifest.Name = GAMENAME_UNNAMED
	}

	var err error
	game.watcher, err = spawnWatcher()
	watchDirectory(game.watcher, dataDirectory())

	game.window, err = NewSDLWindow(
		game.data.Manifest.Name,
		game.data.Settings.Width,
		game.data.Settings.Height,
	)

	if err != nil {
		return nil, errors.New("ERROR! Could not open SDL2 OpenGL window")
	}

	return game, nil
}

func failsafeData() Data {
	return Data{
		Manifest{GAMENAME_LOADING},
		Settings{800, 600, false},
	}
}

func (game *Game) updateSettings() {
	game.window.SetTitle(game.data.Manifest.Name)
	game.window.SetSize(game.data.Settings.Width, game.data.Settings.Height)
}

// Finish cleans up. Closes the games file watcher.
func (game *Game) Finish() {
	game.watcher.Close()
	game.window.Destroy()
}

func spawnWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Warning: Could not create fsnotify watcher")
	}

	return watcher, nil
}

func watchDirectory(watcher *fsnotify.Watcher, root string) {
	watcher.Add(root)
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
	if path.Ext(filename) == ".yml" {
		touched <- filename
	}
}

func (game *Game) handleFileEvents() {
	for {
		select {
		case event := <-game.watcher.Events:
			// [TODO]: Use modification timestamps...
			// For some reason this doesn't work with Vim and *some* files... it CHMOD, RENAME, REMOVE's them rather than WRITE...
			/*
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Modified file:", event.Name)
				game.touched <- event.Name
				}
			*/
			game.touched <- event.Name
		case err := <-game.watcher.Errors:
			log.Println("Error:", err)
		}
	}
}

func (game *Game) consumeAllFileEvents() {
	for game.consumeFileEvents() {
	}
}

func (game *Game) consumeFileEvents() bool {
	select {
	case filename := <-game.touched:
		log.Println("Processing:", filename)
		game.data.LoadYAML(filename)
		game.watcher.Add(filename)
		return true
	default:
		return false
	}
}

func dataDirectory() string {
	return path.Join("data")
}

func watchOrLogError(watcher *fsnotify.Watcher, filename string) error {
	log.Println("Watching file:", filename)
	err := watcher.Add(filename)
	if err != nil {
		log.Println("Warning: Could not watch ", filename, err)
	}
	return err
}

// LoadYAML loads the data form a YAML file.
func (data *Data) LoadYAML(filename string) error {
	log.Println("Loading YAML:", filename)
	rawdata, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Warning: Could not read file.", err)
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
	go game.handleFileEvents()

	for game.running {
		game.consumeAllFileEvents()
		game.updateSettings()
		game.window.Update()
	}

	game.Finish()
}

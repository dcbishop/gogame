package main

import (
	"errors"
	"io/ioutil"
	"log"
	"path"

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

// NewGame constructs a Game.
func NewGame(gamename string) (*Game, error) {
	game := new(Game)
	game.name = gamename
	game.running = true
	game.touched = make(chan string)

	mfilename := path.Join("data", gamename, "manifest.yml")
	sfilename := path.Join("data", gamename, "settings.yml")

	var err error
	game.watcher, err = watchFiles(gamename)

	err = game.data.LoadYAML(mfilename)
	if err != nil {
		log.Println("Warning: Could now load game manifest.", err)
	}

	err = game.data.LoadYAML(sfilename)
	if err != nil {
		log.Println("Warning: Could now load settings.", err)
	}

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

// UpdateSettings applies any changes to the settings.
func (game Game) UpdateSettings() {
	game.window.SetTitle(game.data.Manifest.Name)
	game.window.SetSize(game.data.Settings.Width, game.data.Settings.Height)
}

// Finish cleans up. Closes the games file watcher.
func (game Game) Finish() {
	defer game.watcher.Close()
}

func watchFiles(gamename string) (*fsnotify.Watcher, error) {
	mfilename := manifestFilename(gamename)
	sfilename := settingsFilename(gamename)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Warning: Could not create fsnotify watcher")
		return nil, err
	}

	// [TODO]: Walk the directory for any YAML files.
	// [TODO]: Watch the whole data directory.
	_ = watchOrLogError(watcher, mfilename)
	_ = watchOrLogError(watcher, sfilename)

	return watcher, nil
}

func (game *Game) handleFileEvents() {
	for {
		select {
		case event := <-game.watcher.Events:
			log.Println("Event: ", event)
			// [TODO]: Use modification timestamps...
			// For some reason this doesn't work with vim and *some* files... it CHMOD, RENAME, REMOVE's them rather than WRITE...
			/*
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Modified file:", event.Name)
				game.touched <- event.Name
				}
			*/
			game.touched <- event.Name
		case err := <-game.watcher.Errors:
			log.Println("error:", err)
		}
	}
}

func (game *Game) consumeFileEvents() {
	select {
	case filename := <-game.touched:
		log.Println("Processing", filename)
		game.processFileUpdate(filename)
		game.data.LoadYAML(filename)
	default:
		// Noop
	}
}

func (game *Game) processFileUpdate(filename string) {
	if filename == manifestFilename(game.name) {
		log.Println("Manifest updated")
	}

	if filename == settingsFilename(game.name) {
		log.Println("Settings updated")
	}
}

func manifestFilename(gamename string) string {
	return path.Join("data", gamename, "manifest.yml")
}

func settingsFilename(gamename string) string {
	return path.Join("data", gamename, "settings.yml")
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
func (game Game) Run() {
	go game.handleFileEvents()

	for game.running {
		game.consumeFileEvents()
		//game.UpdateSettings()
		//game.window.Update()
	}
}

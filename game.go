package main

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v1"
)

// Game contains the entire game.
type Game struct {
	window   Window
	opts     Settings
	manifest Manifest
}

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
func NewGame(mfilename string, sfilename string) (*Game, error) {
	game := new(Game)

	var err error
	game.manifest, err = LoadManifest(mfilename)
	if err != nil {
		log.Println("Warning: Could now load game manifest.", err)
	}

	game.opts, err = LoadSettings(sfilename)
	if err != nil {
		log.Println("Warning: Could now load settings.", err)
	}

	game.window, err = NewSDLWindow(
		game.manifest.Name,
		game.opts.Width,
		game.opts.Height,
	)

	if err != nil {
		return nil, errors.New("ERROR! Could not open SDL2 OpenGL window")
	}

	return game, nil
}

// FailsafeSettings returns the default Settings.
func FailsafeSettings() Settings {
	return Settings{800, 600, false}
}

func FailsafeManifest() Manifest {
	return Manifest{"Unnamed"}
}

// LoadSettings loads the Settings form a YAML file.
func LoadSettings(filename string) (Settings, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return FailsafeSettings(), err
	}

	opts, err := parseSettings(data)

	return opts, err
}

func parseSettings(data []byte) (Settings, error) {
	opts := FailsafeSettings()

	err := yaml.Unmarshal(data, &opts)
	if err != nil {
		log.Println("Could not parse settings YAML.")
		return FailsafeSettings(), err
	}

	return opts, nil
}

// LoadManifest loads the manifest file from YAML.
func LoadManifest(filename string) (Manifest, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return FailsafeManifest(), err
	}

	manifest, err := parseManifest(data)

	return manifest, nil
}

func parseManifest(data []byte) (Manifest, error) {
	manifest := FailsafeManifest()

	err := yaml.Unmarshal(data, &manifest)
	if err != nil {
		log.Println("Could not parse manifest YAML.")
		return FailsafeManifest(), err
	}

	return manifest, nil
}

// Run begins the game. Shows the Window, enters the main loop, etc...
func (game Game) Run() {
	game.window.Run()
}

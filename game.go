package main

import (
	"errors"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v1"
)

// Game contains the entire game.
type Game struct {
	window Window
	opts   GameOptions
}

// GameOptions holds options for the game.
type GameOptions struct {
	Name       string
	Width      int
	Height     int
	Fullscreen bool
}

// NewGame constructs a Game.
func NewGame(filename string) *Game {
	game := new(Game)
	var err error
	game.opts, err = LoadOptions(filename)

	if err != nil {
		fmt.Println("Error occured loading game options.")
	}

	game.window = NewSDLWindow(
		game.opts.Name,
		game.opts.Width,
		game.opts.Height,
	)

	return game
}

// FailsafeOptions returns the default GameOptions.
func FailsafeOptions() GameOptions {
	return GameOptions{"Unnamed", 800, 600, false}
}

// LoadOptions loads the GameOptions form a YAML file.
func LoadOptions(filename string) (GameOptions, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return FailsafeOptions(), errors.New("Could not read file.")
	}

	opts, err := parseOptions(data)

	return opts, err
}

func parseOptions(data []byte) (GameOptions, error) {
	opts := FailsafeOptions()

	err := yaml.Unmarshal(data, &opts)
	if err != nil {
		return FailsafeOptions(), errors.New("Could not parse YAML.")
	}

	return opts, nil
}

// Run begins the game. Shows the Window, enters the main loop, etc...
func (game Game) Run() {
	game.window.Run()
}

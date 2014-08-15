package main

import "log"

func main() {
	log.Println("Running...")

	game, err := NewGame()
	if err != nil {
		log.Panic("ERROR: ", err.Error())
	}

	game.Run()
}

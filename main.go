package main

import "log"

func main() {
	log.Println("Running...")

	window, err := NewSDLWindow()
	if err != nil {
		log.Println("ERROR: ", err.Error())
	}

	game := NewGame()

	game.SetDataDirectory(dataDirectory())
	game.SetWindow(window)
	game.Run()
}

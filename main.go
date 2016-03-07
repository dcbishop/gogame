package main

import (
	"log"
	"runtime"
)

func main() {
	log.Println("Running...")

	runtime.LockOSThread()

	window, err := NewSDLWindow()
	if err != nil {
		log.Println("ERROR: ", err.Error())
	}

	game := NewGame()

	game.SetDataDirectory(DataDirectory())
	game.SetWindow(window)
	game.Run()
}

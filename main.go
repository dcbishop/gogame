package main

import "log"

func main() {
	log.Println("Running...")

	game, err := NewGame("testgame")
	if err != nil {
		log.Panic("ERROR: ", err.Error())
	}

	game.Run()
	defer game.Finish()
	//time.Sleep(50 * time.Second)
}

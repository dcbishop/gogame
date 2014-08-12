package main

import (
	"fmt"
)

func main() {
	fmt.Println("Running...")
	game := NewGame("game.yml")
	game.Run()
}

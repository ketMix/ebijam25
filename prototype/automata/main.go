package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	GridWidth    = 100
	GridHeight   = 100
)

func main() {
	game := NewGame()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Cellular Automata Playground")
	ebiten.SetTPS(30)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}

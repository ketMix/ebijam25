package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/game"
)

func main() {
	// It's a great game we've developed here...!!!
	g := &game.Game{}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

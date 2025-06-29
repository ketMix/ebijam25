package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/game"
	"github.com/ketMix/ebijam25/internal/transitions"
	"github.com/ketMix/ebijam25/stuff"
)

func main() {
	// It's a great game we've developed here...!!!
	g := game.NewGame(true)

	tm := &transitions.Manager{}
	g.Managers.Add(tm)
	tm.Add(transitions.NewFade(60*4, true))
	tm.Add(transitions.NewFade(60*4, false))

	if err := stuff.LoadImages(); err != nil {
		panic(err)
	}

	if err := stuff.LoadNames(); err != nil {
		panic(err)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

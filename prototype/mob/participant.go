package main

import "github.com/hajimehoshi/ebiten/v2"

type Participant interface {
	Draw(screen *ebiten.Image)
	Update(participants []Participant)
}

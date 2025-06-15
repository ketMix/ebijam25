package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{32, 0, 32, 255})
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

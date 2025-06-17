package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/client"
)

type Game struct {
	Managers Managers
	client   client.Game
}

func NewGame(localGame bool) *Game {
	g := &Game{}

	g.client.Setup(localGame)

	return g
}

func (g *Game) Update() error {
	g.Managers.Update()
	if err := g.client.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{32, 0, 32, 255})
	g.Managers.Draw(screen)
	g.client.Draw(screen)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	g.client.Layout(ow, oh)
	return ow, oh
}

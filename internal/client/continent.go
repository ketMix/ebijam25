package client

import (
	"github.com/hajimehoshi/ebiten/v2"
)

var fiefImages []*ebiten.Image

func (g *Game) DrawContinent(screen *ebiten.Image) {
	if g.Continent == nil || g.Continent.Fiefs == nil {
		return
	}
	g.DrawFiefs(screen)
}

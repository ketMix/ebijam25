package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
)

func (g *Game) MakeFiefImages(sneed int) {
	fiefs := make([][]*ebiten.Image, world.FiefSpan)
	for i := range fiefs {
		fiefs[i] = make([]*ebiten.Image, world.FiefSpan)
		for j := range fiefs[i] {
			img := ebiten.NewImage(world.FiefSize, world.FiefSize)
			fsneed := sneed + i + j
			randomColor := func() color.Color {
				return color.RGBA{
					R: uint8((fsneed*31 + 17) % 256),
					G: uint8((fsneed*37 + 29) % 256),
					B: uint8((fsneed*41 + 43) % 256),
					A: 50,
				}
			}
			img.Fill(randomColor())
			fiefs[i][j] = img
		}
	}
	g.fiefImages = fiefs
}

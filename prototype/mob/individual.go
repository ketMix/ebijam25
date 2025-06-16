package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Individual struct {
	name string
	x, y float64
}

func (i *Individual) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(i.x, i.y)
	img := images["chump"]
	opts.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
	screen.DrawImage(img, opts)
}

func (i *Individual) Update(participants []Participant) {
	radius := len(participants) * 2
	// Try to spread ourself out from other participants.
	for _, p := range participants {
		if ind, ok := p.(*Individual); ok && ind != i {
			dx := i.x - ind.x
			dy := i.y - ind.y
			dist := dx*dx + dy*dy
			if dist < float64(radius)*4 {
				// Move away from the other individual.
				i.x += dx / float64(radius)
				i.y += dy / float64(radius)
			} else if dist > float64(radius*8) {
				// Move towards the other individual.
				i.x -= dx / float64(radius)
				i.y -= dy / float64(radius)
			}
		}
	}
}

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Individual struct {
	image *ebiten.Image
	name  string
	x, y  float64
}

func NewIndividual(name string, x, y float64) *Individual {
	return &Individual{
		name:  name,
		x:     x,
		y:     y,
		image: images["chump"],
	}
}

func (i *Individual) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(i.x, i.y)
	opts.GeoM.Translate(-float64(i.image.Bounds().Dx())/2, -float64(i.image.Bounds().Dy())/2)
	screen.DrawImage(i.image, opts)
}

func (i *Individual) Update(participants []Participant) {
	radius := float64(len(participants)) * 8
	// Push away from others if too close and pull together if too far away.
	vx := 0.0
	vy := 0.0
	for _, p := range participants {
		if ind, ok := p.(*Individual); ok && ind != i {
			dx := i.x - ind.x
			dy := i.y - ind.y
			dist := dx*dx + dy*dy
			if dist < radius { // Too close, push away.
				vx += dx / dist * radius * 0.1
				vy += dy / dist * radius * 0.1
			} else if dist > radius { // Too far, pull together.
				vx -= dx / dist * radius * 0.01
				vy -= dy / dist * radius * 0.01
			}
		}
	}
	if vx > 2 {
		vx = 2
	}
	if vx < -2 {
		vx = -2
	}
	if vy > 2 {
		vy = 2
	}
	if vy < -2 {
		vy = -2
	}
	i.x += vx
	i.y += vy
}

package main

import "github.com/hajimehoshi/ebiten/v2"

type Structure struct {
	name  string
	x, y  float64
	timer int
	rate  int
}

func (s *Structure) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(s.x, s.y)
	img := images["village"]
	opts.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
	screen.DrawImage(img, opts)
}

func (s *Structure) Update() {
	s.timer++
	if s.timer > s.rate {
		s.timer = 0
		eventBus.Publish(&EventProduce{
			structure: s,
			individual: &Individual{
				name: "chump",
				x:    s.x,
				y:    s.y,
			},
		})
	}
}

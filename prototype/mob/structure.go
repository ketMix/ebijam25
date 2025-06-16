package main

import "github.com/hajimehoshi/ebiten/v2"

type Structure struct {
	name     string
	x, y     float64
	timer    int
	rate     int
	failures int
}

func (s *Structure) Draw(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(s.x, s.y)
	var img *ebiten.Image
	if s.failures > 1 {
		img = images["village-dead"]
	} else {
		img = images["village"]
	}
	opts.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
	screen.DrawImage(img, opts)
}

func (s *Structure) Update(participants []Participant) {
	if s.failures > 5 {
		// Kill off the structure!
		return
	}
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

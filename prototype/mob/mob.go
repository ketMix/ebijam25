package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Individual struct {
	name string
	x, y float64 // x, y in range of -1 to 1, indicating distance from center of mob.
}

type Mob struct {
	individuals      []*Individual
	structures       []*Structure // mobile structures.
	x, y             float64
	targetX, targetY float64
}

func (m *Mob) AddIndividual(ind *Individual) {
	m.individuals = append(m.individuals, ind)
}

func (m *Mob) AddStructure(s *Structure) {
	m.structures = append(m.structures, s)
}

func (m *Mob) Draw(screen *ebiten.Image) {
	radius := len(m.individuals) * 2
	vector.StrokeCircle(screen, float32(m.x), float32(m.y), float32(radius)/2, 1, color.NRGBA{255, 0, 255, 255}, true)

	// Draw structures.
	for _, s := range m.structures {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(s.x, s.y)
		img := images["mobile-village"]
		opts.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
		screen.DrawImage(img, opts)
	}

	for _, ind := range m.individuals {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(ind.x, ind.y)
		img := images["chump"]
		opts.GeoM.Translate(-float64(img.Bounds().Dx())/2, -float64(img.Bounds().Dy())/2)
		screen.DrawImage(img, opts)
	}
}

func (m *Mob) Update() {
	// Move towards targetX and targetY.
	if m.x != m.targetX || m.y != m.targetY {
		angleToTarget := math.Atan2(m.targetY-m.y, m.targetX-m.x)
		dx := math.Cos(angleToTarget) * 2
		dy := math.Sin(angleToTarget) * 2
		m.x += dx
		m.y += dy
		// If close enough, snap to target.
		if math.Abs(m.x-m.targetX) < 1 && math.Abs(m.y-m.targetY) < 1 {
			m.x = m.targetX
			m.y = m.targetY
		}
	}

	// Spread out individuals but pull them towards the mob's x and y.
	radius := float64(len(m.individuals) * 2)
	for _, ind := range m.individuals {
		// Move towards mob center, faster based on distance.
		dx := m.x - ind.x
		dy := m.y - ind.y
		dist := (dx*dx + dy*dy)
		if dist > radius/2 {
			// Move towards the mob center.
			ind.x += dx * 0.01
			ind.y += dy * 0.01
		}
		// Space from other individuals.
		for _, other := range m.individuals {
			if other != ind {
				dx := other.x - ind.x
				dy := other.y - ind.y
				dist := (dx*dx + dy*dy)
				if dist < 4 {
					// Move away from the other individual.
					ind.x -= dx * 4
					ind.y -= dy * 4
				}
			}
		}
	}

	// Spread out structures but pull them towards the mob's x and y.
	for _, s := range m.structures {
		// Move towards mob center, faster based on distance.
		dx := m.x - s.x
		dy := m.y - s.y
		dist := (dx*dx + dy*dy)
		if dist > 16 {
			// Move towards the mob center.
			s.x += dx * 0.01
			s.y += dy * 0.01
		}
	}
	// Spread out structures from each other.
	for i, s := range m.structures {
		for j, other := range m.structures {
			if i != j {
				dx := other.x - s.x
				dy := other.y - s.y
				dist := (dx*dx + dy*dy)
				if dist < 16 {
					// Move away from the other structure.
					s.x -= dx * 0.1
					s.y -= dy * 0.1
				}
			}
		}
		s.Update()
	}
}

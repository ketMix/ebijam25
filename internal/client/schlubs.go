package client

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
)

const MaxSchlubs = 1000
const SchlubRadius = 4.0
const SchlubDiameter = SchlubRadius * 2.0
const SchlubWidth = int(SchlubDiameter * 16.0)

// we can potentially alternate between updating and just applying velocities with some damping
// right now everything explodes when settling at the center if we don't update every tick
const SchlubUpdateTick = 1

// single schlub
type Schlub struct {
	world.Schlub
	VX, VY float64
	Color  [4]float32
}

type Schlubs struct {
	schlubs     []Schlub
	mob         *world.Mob
	time        float64
	tick        int
	schlubImage *ebiten.Image
}

func getSchlubImage() *ebiten.Image {
	// draw the schlub image
	schlumbImg := ebiten.NewImage(SchlubWidth, SchlubWidth)
	for y := range SchlubWidth {
		for x := range SchlubWidth {
			width := float64(SchlubWidth) / 2.0
			dx := float64(x) - width
			dy := float64(y) - width
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < width {
				// gradient from center to edge
				alpha := uint8(255 * (1 - dist/width))
				schlumbImg.Set(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}
	return schlumbImg
}

func NewSchlubSystem(mob *world.Mob) *Schlubs {
	s := &Schlubs{
		mob:         mob,
		schlubs:     make([]Schlub, 0, len(mob.Schlubs)),
		schlubImage: getSchlubImage(),
	}

	schlubCount := len(s.mob.Schlubs)
	if schlubCount == 0 {
		return s
	}

	s.schlubs = s.schlubs[:0]
	mobRadius := s.mob.Radius()
	centerX, centerY := s.mob.X, s.mob.Y

	// spiral the initial schlubs to prevent explosions
	for i := range schlubCount {
		spiralIndex := float64(i)
		angle := spiralIndex * 2.37 // goldn rat
		distance := math.Sqrt(spiralIndex) * SchlubDiameter * 0.8
		maxDistance := mobRadius - SchlubRadius
		if distance > maxDistance {
			distance = maxDistance
		}

		x := centerX + distance*math.Cos(angle)
		y := centerY + distance*math.Sin(angle)

		s.schlubs = append(s.schlubs, Schlub{
			Schlub: world.Schlub{
				X: x,
				Y: y,
			},
			Color: [4]float32{
				// TODO: Mob (Player) color?
				//       Color from type?
				1.0,
				1.0,
				1.0,
				1.0,
			},
		})
	}
	return s
}

// adds existing schlub from another mob
func (s *Schlubs) PersuadeSchlub(gullySchlub *Schlub) {
	if len(s.schlubs) >= MaxSchlubs {
		return // ??
	}

	s.schlubs = append(s.schlubs, Schlub{
		Schlub: world.Schlub{
			X: gullySchlub.X,
			Y: gullySchlub.Y,
		},
		Color: [4]float32{1.0, 1.0, 1.0, 1.0},
	})
	return
}

// remove schlubs roughly around the given position
func (s *Schlubs) LoseSchlubs(x, y float64, count int) {
	// TODO: this method
}

func (s *Schlubs) Update() {
	s.tick++
	s.time++

	if s.tick >= SchlubUpdateTick {
		s.tick = 0
	} else {
		// If not updating, just apply velocities to keep movement smooth
		for i := range s.schlubs {
			p := &s.schlubs[i]
			p.X += p.VX
			p.Y += p.VY
		}
		return
	}

	mobRadius := s.mob.Radius()
	centerX, centerY := s.mob.X, s.mob.Y

	// Phase 1: Apply center attraction and gentle movement
	for i := range s.schlubs {
		p := &s.schlubs[i]

		// Center attraction force
		dx := centerX - p.X
		dy := centerY - p.Y
		distToCenter := math.Sqrt(dx*dx + dy*dy)

		// Strong center attraction with distance falloff
		centerForce := 0.1
		if distToCenter > 0.1 {
			p.VX += (dx / distToCenter) * centerForce
			p.VY += (dy / distToCenter) * centerForce
		}

		// Gentle circular motion around center
		tangentX := -dy
		tangentY := dx
		if distToCenter > 0.1 {
			tangentX /= distToCenter
			tangentY /= distToCenter
		}

		orbitalForce := 0.02 * math.Sin(s.time+float64(i)*0.1)
		p.VX += tangentX * orbitalForce
		p.VY += tangentY * orbitalForce

		// Light damping
		p.VX *= 0.9
		p.VY *= 0.9
	}

	// Phase 2: Collision resolution (prevent overlap)
	for i := range s.schlubs {
		for j := i + 1; j < len(s.schlubs); j++ {
			p1 := &s.schlubs[i]
			p2 := &s.schlubs[j]

			dx := p2.X - p1.X
			dy := p2.Y - p1.Y
			dist := math.Sqrt(dx*dx + dy*dy)
			minDist := SchlubDiameter

			if dist < minDist && dist > 0.1 {
				overlap := minDist - dist

				// Normalize direction
				nx := dx / dist
				ny := dy / dist

				// Add separation velocity
				repulsionVel := overlap * 0.35
				p1.VX -= nx * repulsionVel
				p1.VY -= ny * repulsionVel
				p2.VX += nx * repulsionVel
				p2.VY += ny * repulsionVel
			}
		}
	}

	// Phase 3: Apply velocities and constrain to mob area
	for i := range s.schlubs {
		p := &s.schlubs[i]

		// Apply velocity
		p.X += p.VX
		p.Y += p.VY

		// Keep particles within mob bounds (with some padding for particle radius)
		dx := p.X - centerX
		dy := p.Y - centerY
		distFromCenter := math.Sqrt(dx*dx + dy*dy)
		maxDist := mobRadius - SchlubRadius

		if distFromCenter > maxDist {
			// Clamp to boundary
			if distFromCenter > 0.1 {
				p.X = centerX + (dx/distFromCenter)*maxDist
				p.Y = centerY + (dy/distFromCenter)*maxDist
			}
			// Stop velocity in direction away from center
			if p.VX*dx+p.VY*dy > 0 {
				p.VX *= 0.1
				p.VY *= 0.1
			}
		}
	}
}

func (s *Schlubs) Draw(screen *ebiten.Image) {
	for _, p := range s.schlubs {
		s.drawSchlub(screen, p)
	}
}

func (ps *Schlubs) drawSchlub(screen *ebiten.Image, p Schlub) {
	width := float64(SchlubWidth)
	scale := SchlubRadius / width

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-width, -width)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(p.X, p.Y)
	screen.DrawImage(ps.schlubImage, op)
}

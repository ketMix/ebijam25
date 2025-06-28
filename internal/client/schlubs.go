package client

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
	"github.com/ketMix/ebijam25/stuff"
)

const (
	SchlubRadius   = 4.0
	SchlubDiameter = SchlubRadius * 2.0

	// Physics constants
	CenterAttraction = 1.0
	OrbitalForce     = 0.35
	Damping          = 0.5
	RepulsionVel     = 0.25
	BoundaryDamping  = 0.01
)

// Schlub represents a single constituent/denizen of a mob using polar coordinates
type Schlub struct {
	world.Schlub
	Distance     float64 // Distance from mob center
	Angle        float64 // Angle from mob center (radians)
	VDistance    float64 // Radial velocity (toward/away from center)
	VAngle       float64 // Angular velocity (orbital motion)
	GridX, GridY int
}

type polarGrid struct {
	angleSlices int     // Number of angular divisions
	radialBands float64 // Width of each radial band
	cells       map[uint64][]int
}

func newPolarGrid() *polarGrid {
	return &polarGrid{
		angleSlices: 32,                   // Divide circle into 32 slices
		radialBands: SchlubDiameter * 8.0, // Each radial band is 2 schlub diameters wide
		cells:       make(map[uint64][]int),
	}
}

func (g *polarGrid) getCoords(angle, distance float64) (int, int) {
	// Normalize angle to [0, 2π]
	normalizedAngle := math.Mod(angle, 2*math.Pi)
	if normalizedAngle < 0 {
		normalizedAngle += 2 * math.Pi
	}

	angleIndex := int(normalizedAngle * float64(g.angleSlices) / (2 * math.Pi))
	radialIndex := int(distance / g.radialBands)
	return angleIndex, radialIndex
}

func (g *polarGrid) packCoords(angleIdx, radialIdx int) uint64 {
	return uint64(uint32(angleIdx))<<32 | uint64(uint32(radialIdx))
}

func (g *polarGrid) insert(idx int, angle, distance float64) {
	aIdx, rIdx := g.getCoords(angle, distance)
	key := g.packCoords(aIdx, rIdx)
	g.cells[key] = append(g.cells[key], idx)
}

func (g *polarGrid) getNeighborIndices(angle, distance float64) []int {
	neighbors := make([]int, 0, 32)
	aIdx, rIdx := g.getCoords(angle, distance)

	// Check neighboring cells in polar grid
	// Check current and adjacent radial bands
	for dr := -1; dr <= 1; dr++ {
		// Check current and adjacent angular slices
		for da := -1; da <= 1; da++ {
			// Handle wrap-around for angular index
			neighborAngleIdx := (aIdx + da + g.angleSlices) % g.angleSlices
			neighborRadialIdx := rIdx + dr

			if neighborRadialIdx >= 0 {
				key := g.packCoords(neighborAngleIdx, neighborRadialIdx)
				if indices, exists := g.cells[key]; exists {
					neighbors = append(neighbors, indices...)
				}
			}
		}
	}
	return neighbors
}

type Schlubs struct {
	schlubs         []*Schlub
	mob             *world.Mob
	time            float64
	playerImage     *ebiten.Image
	vagrantImage    *ebiten.Image
	monkImage       *ebiten.Image
	warriorImage    *ebiten.Image
	grid            *polarGrid
	outerSchlubKind world.SchlubID
	outerRadius     float64

	toRemove []int
}

func getSchlubImage(kind int) *ebiten.Image {
	var img *ebiten.Image
	if kind == int(world.SchlubKindMonk) {
		img = stuff.GetImage("monke")
	} else if kind == int(world.SchlubKindWarrior) {
		img = stuff.GetImage("warrior")
	} else if kind == int(world.SchlubKindPlayer) {
		img = stuff.GetImage("player")
	}
	if img == nil {
		img = stuff.GetImage("vagrant") // Default to vagrant if unknown kind
	}
	return img
}

func NewSchlubs(mob *world.Mob) *Schlubs {
	s := &Schlubs{
		mob:             mob,
		schlubs:         make([]*Schlub, 0, len(mob.Schlubs)),
		playerImage:     getSchlubImage(int(world.SchlubKindPlayer)),
		vagrantImage:    getSchlubImage(int(world.SchlubKindVagrant)),
		monkImage:       getSchlubImage(int(world.SchlubKindMonk)),
		warriorImage:    getSchlubImage(int(world.SchlubKindWarrior)),
		grid:            newPolarGrid(),
		toRemove:        make([]int, 0, 32),
		outerSchlubKind: world.SchlubKindVagrant, // Start with vagrant
	}

	if len(s.mob.Schlubs) == 0 {
		return s
	}

	mobRadius := s.mob.Radius()

	for i, schlub := range mob.Schlubs {
		spiralIndex := float64(i)
		// Golden angle for even distribution
		angle := spiralIndex * 2.39

		baseDistance := math.Sqrt(spiralIndex) * SchlubDiameter * 0.8
		jitter := (rand.Float64() - 0.5) * SchlubRadius * 0.5
		distance := baseDistance + jitter

		maxDistance := mobRadius - SchlubRadius
		if distance > maxDistance {
			distance = maxDistance
			angle += rand.Float64() * 0.5
		}

		s.schlubs = append(s.schlubs, &Schlub{
			Schlub: world.Schlub{
				ID: schlub,
			},
			Distance: distance,
			Angle:    angle,
		})
	}
	s.updateRadius()
	return s
}

// Helper function to get Cartesian position for a schlub
func (s *Schlubs) getCartesian(schlub *Schlub) (float64, float64) {
	x := s.mob.X + schlub.Distance*math.Cos(schlub.Angle)
	y := s.mob.Y + schlub.Distance*math.Sin(schlub.Angle)
	return x, y
}

func (s *Schlubs) Swap() {
	if s.outerSchlubKind == world.SchlubKindVagrant {
		s.outerSchlubKind = world.SchlubKindMonk
	} else if s.outerSchlubKind == world.SchlubKindMonk {
		s.outerSchlubKind = world.SchlubKindWarrior
	} else {
		s.outerSchlubKind = world.SchlubKindVagrant
	}

	s.updateRadius()
}

func (s *Schlubs) updateRadius() {
	// Update inner and outer radius based on mob size
	totalSchlubs := len(s.schlubs)
	if totalSchlubs == 0 {
		s.outerRadius = 0
		return
	}

	innerSchlubs := 0
	for _, schlub := range s.schlubs {
		if schlub.ID.KindID() == int(s.outerSchlubKind) {
			innerSchlubs++
		}
	}
	// Total radius
	radiusSchlubs := math.Sqrt(float64(totalSchlubs) / math.Pi)
	schlubWidth := s.mob.Spread() / radiusSchlubs
	innerRadius := schlubWidth * (float64(innerSchlubs) / float64(totalSchlubs))
	s.outerRadius = innerRadius + SchlubDiameter*4.0
}

func (s *Schlubs) getSchlubColor() color.NRGBA {
	return s.mob.Color
}

// PersuadeSchlub adds an existing schlub from another mob
func (s *Schlubs) PersuadeSchlubs(gullySchlubs []*world.Schlub) {
	if len(s.schlubs) >= world.MaxSchlubsPerMob {
		return
	}

	if len(gullySchlubs) == 0 {
		return
	}
	for _, g := range gullySchlubs {
		// Convert from Cartesian to polar coordinates
		dx := g.X - s.mob.X
		dy := g.Y - s.mob.Y
		distance := math.Sqrt(dx*dx + dy*dy)
		angle := math.Atan2(dy, dx)

		newSchlub := &Schlub{
			Schlub: world.Schlub{
				ID: g.ID,
			},
			Distance: distance,
			Angle:    angle,
		}
		s.schlubs = append(s.schlubs, newSchlub)
	}
}

func (s *Schlubs) applyForces() {
	for i := range s.schlubs {
		p := s.schlubs[i]

		var minRadius float64
		var maxRadius float64
		if p.ID.KindID() == int(world.SchlubKindPlayer) {
			minRadius = 0
			maxRadius = 0
		} else if p.ID.KindID() == int(s.outerSchlubKind) {
			minRadius = s.outerRadius
		} else {
			// Leave space for the inner player schlub
			minRadius = SchlubDiameter * 4.0
		}
		maxRadius = minRadius + SchlubDiameter*4.0

		// Apply radial forces based on distance constraints
		if p.Distance > maxRadius || p.Distance < minRadius {
			var direction = 1.0
			if p.Distance < minRadius {
				direction = -1.0
			}

			attraction := CenterAttraction * (1 + p.Distance*0.001)
			p.VDistance -= attraction * direction
		}

		// Apply orbital motion
		if p.ID.KindID() == int(s.outerSchlubKind) {
			p.VAngle += OrbitalForce / p.Distance
		} else {
			p.VAngle -= OrbitalForce / p.Distance
		}

		// Apply damping
		p.VDistance *= Damping
		p.VAngle *= Damping
	}
}

// Version with polar grid for better performance with many schlubs:
func (s *Schlubs) resolveCollisions() {
	// Use polar grid for spatial partitioning
	grid := newPolarGrid()

	// Insert all schlubs into polar grid
	for i, schlub := range s.schlubs {
		grid.insert(i, schlub.Angle, schlub.Distance)
	}

	for i, schlub := range s.schlubs {
		neighbors := grid.getNeighborIndices(schlub.Angle, schlub.Distance)
		outerSchlub := schlub.ID.KindID() == int(s.outerSchlubKind)
		for _, j := range neighbors {
			if j <= i {
				continue
			}

			otherSchlub := s.schlubs[j]
			outerOtherSchlub := otherSchlub.ID.KindID() == int(s.outerSchlubKind)
			angleDiff := otherSchlub.Angle - schlub.Angle
			distSq := schlub.Distance*schlub.Distance +
				otherSchlub.Distance*otherSchlub.Distance -
				2*schlub.Distance*otherSchlub.Distance*math.Cos(angleDiff)

			minDist := SchlubDiameter
			if outerSchlub != outerOtherSchlub {
				// passthrough schlubs
				minDist *= 0.9
			}
			minDistSq := minDist * minDist

			if distSq < minDistSq || distSq < 0.1 {
				dist := math.Sqrt(distSq)
				overlap := minDist - dist

				collisionAngle := math.Atan2(
					otherSchlub.Distance*math.Sin(otherSchlub.Angle)-schlub.Distance*math.Sin(schlub.Angle),
					otherSchlub.Distance*math.Cos(otherSchlub.Angle)-schlub.Distance*math.Cos(schlub.Angle),
				)

				separationForce := overlap * RepulsionVel

				angleToCollision := collisionAngle - schlub.Angle
				schlub.VDistance -= separationForce * math.Cos(angleToCollision)
				schlub.VAngle -= separationForce * math.Sin(angleToCollision) / schlub.Distance

				angleFromCollision := collisionAngle - otherSchlub.Angle + math.Pi
				otherSchlub.VDistance -= separationForce * math.Cos(angleFromCollision)
				otherSchlub.VAngle -= separationForce * math.Sin(angleFromCollision) / otherSchlub.Distance
			}
		}
	}
}

func (s *Schlubs) integrateAndConstrain() {
	maxDist := s.mob.Radius() - SchlubRadius

	for i := range s.schlubs {
		p := s.schlubs[i]

		// Update polar coordinates
		p.Distance += p.VDistance
		p.Angle += p.VAngle

		// Normalize angle to [0, 2π]
		p.Angle = math.Mod(p.Angle, 2*math.Pi)
		if p.Angle < 0 {
			p.Angle += 2 * math.Pi
		}

		if p.Distance > maxDist {
			p.Distance = maxDist
			if p.VDistance > 0 {
				p.VDistance *= -BoundaryDamping
			}
		}

		// Prevent negative distance
		if p.Distance < SchlubRadius {
			p.Distance = SchlubRadius
			if p.VDistance < 0 {
				p.VDistance *= -BoundaryDamping
			}
		}
	}
}

func (s *Schlubs) UpdateMob(serverRate float64) {
	if s.mob == nil || s.mob.TargetX == s.mob.X || s.mob.TargetY == s.mob.Y {
		return
	}
	m := s.mob
	angleToTarget := math.Atan2(m.TargetY-m.Y, m.TargetX-m.X)
	dx := math.Cos(angleToTarget)
	dy := math.Sin(angleToTarget)
	speed := serverRate * (serverRate / ebiten.ActualTPS())
	x := m.X + dx*speed
	y := m.Y + dy*speed

	if math.Abs(x-m.TargetX) < speed && math.Abs(y-m.TargetY) < speed {
		x = m.TargetX
		x = m.TargetY
	}
	m.X = x
	m.Y = y
}

func (s *Schlubs) Update(serverRate float64) {
	s.updateRadius()
	s.UpdateMob(serverRate)

	s.applyForces()
	s.resolveCollisions()
	s.integrateAndConstrain()
}

func (s *Schlubs) Draw(screen *ebiten.Image) {
	for _, p := range s.schlubs {
		op := &ebiten.DrawImageOptions{}
		x, y := s.getCartesian(p)
		op.GeoM.Translate(x, y)
		color := s.getSchlubColor()
		op.ColorScale.ScaleWithColor(color)

		if p.ID.KindID() == int(world.SchlubKindMonk) {
			screen.DrawImage(s.monkImage, op)
		} else if p.ID.KindID() == int(world.SchlubKindWarrior) {
			screen.DrawImage(s.warriorImage, op)
		} else if p.ID.KindID() == int(world.SchlubKindPlayer) {
			screen.DrawImage(s.playerImage, op)
		} else {
			screen.DrawImage(s.vagrantImage, op)
		}
	}
}

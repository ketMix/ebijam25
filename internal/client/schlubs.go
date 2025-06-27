package client

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ketMix/ebijam25/internal/world"
	"github.com/ketMix/ebijam25/stuff"
)

const (
	MaxSchlubs       = 1000
	SchlubRadius     = 4.0
	SchlubDiameter   = SchlubRadius * 2.0
	SchlubWidth      = SchlubDiameter * 16.0
	SchlubUpdateTick = 1 // Can be increased for performance

	// Physics constants
	CenterAttraction = 0.15
	OrbitalForce     = 0.02
	Damping          = 0.9
	RepulsionVel     = 0.4
	BoundaryDamping  = 0.1
)

// Schlub represents a single constituent/denizen of a mob.
type Schlub struct {
	world.Schlub
	VX, VY       float64
	GridX, GridY int
}

// spatialGrid for efficient collision detection
type spatialGrid struct {
	cellSize float64
	cells    map[uint64][]int
}

func newSpatialGrid() *spatialGrid {
	return &spatialGrid{
		cellSize: SchlubDiameter * 2,
		cells:    make(map[uint64][]int),
	}
}

func (g *spatialGrid) clear() {
	for k := range g.cells {
		delete(g.cells, k)
	}
}

func (g *spatialGrid) getCoords(x, y float64) (int, int) {
	return int(x / g.cellSize), int(y / g.cellSize)
}

func (g *spatialGrid) packCoords(gx, gy int) uint64 {
	// Pack grid coords into single uint64
	return uint64(uint32(gx))<<32 | uint64(uint32(gy))
}

func (g *spatialGrid) insert(idx int, x, y float64) (int, int) {
	gx, gy := g.getCoords(x, y)
	key := g.packCoords(gx, gy)
	g.cells[key] = append(g.cells[key], idx)
	return gx, gy
}

func (g *spatialGrid) getNeighborIndices(gx, gy int) []int {
	neighbors := make([]int, 0, 32)

	// Check 3x3 grid around position
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			key := g.packCoords(gx+dx, gy+dy)
			if indices, exists := g.cells[key]; exists {
				neighbors = append(neighbors, indices...)
			}
		}
	}
	return neighbors
}

type Schlubs struct {
	schlubs         []Schlub
	mob             *world.Mob
	time            float64
	tick            int
	vagrantImage    *ebiten.Image
	monkImage       *ebiten.Image
	warriorImage    *ebiten.Image
	grid            *spatialGrid
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
	}
	if img == nil {
		img = stuff.GetImage("vagrant") // Default to vagrant if unknown kind
	}
	halfWidth := SchlubWidth / 2.0

	for y := range int(SchlubWidth) {
		for x := range int(SchlubWidth) {
			dx := float64(x) - halfWidth
			dy := float64(y) - halfWidth
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < halfWidth {
				// Smoother gradient with quadratic falloff
				t := dist / halfWidth
				alpha := uint8(255 * (1 - t*t))

				// Add subtle noise for visual interest
				if rand.Intn(10) < 3 {
					alpha = uint8(math.Min(255, float64(alpha)+5))
				}

				img.Set(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}
	return img
}

func NewSchlubs(mob *world.Mob) *Schlubs {
	s := &Schlubs{
		mob:             mob,
		schlubs:         make([]Schlub, 0, len(mob.Schlubs)),
		vagrantImage:    getSchlubImage(int(world.SchlubKindVagrant)),
		monkImage:       getSchlubImage(int(world.SchlubKindMonk)),
		warriorImage:    getSchlubImage(int(world.SchlubKindWarrior)),
		grid:            newSpatialGrid(),
		toRemove:        make([]int, 0, 32),
		outerSchlubKind: world.SchlubKindVagrant, // Start with vagrant
	}

	if len(s.mob.Schlubs) == 0 {
		return s
	}

	mobRadius := s.mob.Radius()
	centerX, centerY := s.mob.X, s.mob.Y

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

		x := centerX + distance*math.Cos(angle)
		y := centerY + distance*math.Sin(angle)
		s.schlubs = append(s.schlubs, Schlub{
			Schlub: world.Schlub{
				ID: schlub,
				X:  x,
				Y:  y,
			},
		})
	}
	s.UpdateRadius()
	return s
}

func (s *Schlubs) Swap() {
	if s.outerSchlubKind == world.SchlubKindVagrant {
		s.outerSchlubKind = world.SchlubKindMonk
	} else if s.outerSchlubKind == world.SchlubKindMonk {
		// s.outerSchlubKind = world.SchlubKindWarrior
		s.outerSchlubKind = world.SchlubKindVagrant
	} else {
		s.outerSchlubKind = world.SchlubKindVagrant
	}

	s.UpdateRadius()
}

func (s *Schlubs) UpdateRadius() {
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
	s.outerRadius = max(math.Sqrt(float64(innerSchlubs)*math.Pi)*SchlubRadius*(float64(totalSchlubs-innerSchlubs)/float64(innerSchlubs)), SchlubWidth/2)
}

func (s *Schlubs) getSchlubColor() [4]float32 {
	// TODO: Use mob color or type-based color
	// For now, slight variations in brightness
	brightness := 0.85 + rand.Float32()*0.15
	return [4]float32{brightness, brightness, brightness, 1.0}
}

// PersuadeSchlub adds an existing schlub from another mob
func (s *Schlubs) PersuadeSchlubs(gullySchlubs []*world.Schlub) {
	if len(s.schlubs) >= MaxSchlubs {
		return
	}

	if len(gullySchlubs) == 0 {
		return
	}
	for _, g := range gullySchlubs {
		newSchlub := Schlub{
			Schlub: world.Schlub{
				ID: g.ID,
				X:  g.X,
				Y:  g.Y,
			},
		}
		s.schlubs = append(s.schlubs, newSchlub)
	}
}

// LoseSchlubs removes schlubs near the given position
func (s *Schlubs) LoseSchlubs(x, y float64, count int) []Schlub {
	if len(s.schlubs) == 0 || count <= 0 {
		return nil
	}

	// Calculate distances to all schlubs
	type distSchlub struct {
		dist  float64
		index int
	}

	distances := make([]distSchlub, len(s.schlubs))
	for i, schlub := range s.schlubs {
		dx := schlub.X - x
		dy := schlub.Y - y
		distances[i] = distSchlub{
			dist:  dx*dx + dy*dy, // Use squared distance
			index: i,
		}
	}

	// Sort by distance
	sort.Slice(distances, func(i, j int) bool {
		return distances[i].dist < distances[j].dist
	})

	// Remove the closest schlubs
	removeCount := min(count, len(s.schlubs))
	removed := make([]Schlub, removeCount)

	// Collect indices to remove (in reverse order for safe removal)
	s.toRemove = s.toRemove[:0]
	for i := 0; i < removeCount; i++ {
		idx := distances[i].index
		removed[i] = s.schlubs[idx]
		s.toRemove = append(s.toRemove, idx)
	}

	// Sort indices in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(s.toRemove)))

	// Remove from slice
	for _, idx := range s.toRemove {
		s.schlubs[idx] = s.schlubs[len(s.schlubs)-1]
		s.schlubs = s.schlubs[:len(s.schlubs)-1]
	}

	// Update mob's schlub list
	s.mob.Schlubs = s.mob.Schlubs[:len(s.schlubs)]

	return removed
}

func (s *Schlubs) Update() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		s.Swap()
	}

	s.tick++
	s.time += 1.0 / 60.0

	if s.tick < SchlubUpdateTick {
		// Just apply velocities for smooth movement
		for i := range s.schlubs {
			p := &s.schlubs[i]
			p.X += p.VX * Damping
			p.Y += p.VY * Damping
		}
		return
	}
	s.tick = 0

	mobRadius := s.mob.Radius()
	centerX, centerY := s.mob.X, s.mob.Y

	s.applyForces(centerX, centerY)
	s.resolveCollisions()
	s.integrateAndConstrain(centerX, centerY, mobRadius)
}

func (s *Schlubs) applyForces(centerX, centerY float64) {
	for i := range s.schlubs {
		p := &s.schlubs[i]

		var minRadius float64
		var maxRadius float64
		isInner := p.ID.KindID() == int(s.outerSchlubKind)
		if isInner {
			minRadius = 0
			maxRadius = SchlubDiameter
		} else {
			minRadius = s.outerRadius - SchlubDiameter
			maxRadius = s.outerRadius
		}

		// Vector to center
		dx := centerX - p.X
		dy := centerY - p.Y
		distSq := dx*dx + dy*dy
		if distSq > maxRadius*maxRadius || distSq < minRadius*minRadius {
			dist := math.Sqrt(distSq)
			nx := dx / dist
			ny := dy / dist

			var direction float64
			if dist > maxRadius {
				direction = 1.0
			} else {
				direction = -1.0
			}

			attraction := CenterAttraction * (1 + dist*0.001)
			p.VX += nx * attraction * direction
			p.VY += ny * attraction * direction

			// Orbital motion
			tangentX := -ny
			tangentY := nx
			phase := s.time*0.1 + float64(i)*0.1
			orbital := OrbitalForce * math.Sin(phase)
			p.VX += tangentX * orbital
			p.VY += tangentY * orbital
		}

		// Apply damping
		p.VX *= Damping
		p.VY *= Damping
	}
}

func (s *Schlubs) resolveCollisions() {
	// Build spatial grid
	s.grid.clear()
	for i := range s.schlubs {
		gx, gy := s.grid.insert(i, s.schlubs[i].X, s.schlubs[i].Y)
		s.schlubs[i].GridX = gx
		s.schlubs[i].GridY = gy
	}

	// Check collisions using spatial grid
	minDistSq := SchlubDiameter * SchlubDiameter

	for i := range s.schlubs {
		p1 := &s.schlubs[i]
		neighbors := s.grid.getNeighborIndices(p1.GridX, p1.GridY)

		isP1Inner := p1.ID.KindID() == int(s.outerSchlubKind)
		for _, j := range neighbors {
			if j <= i {
				continue
			}

			p2 := &s.schlubs[j]
			isP2Inner := p2.ID.KindID() == int(s.outerSchlubKind)
			if isP1Inner != isP2Inner {
				continue
			}

			dx := p2.X - p1.X
			dy := p2.Y - p1.Y
			distSq := dx*dx + dy*dy

			if distSq < minDistSq && distSq > 0.01 {
				// Resolve collision
				dist := math.Sqrt(distSq)
				overlap := SchlubDiameter - dist

				// Normalized separation
				nx := dx / dist
				ny := dy / dist

				// Velocity-based separation (more stable)
				vel := overlap * RepulsionVel
				p1.VX -= nx * vel
				p1.VY -= ny * vel
				p2.VX += nx * vel
				p2.VY += ny * vel
			}
		}
	}
}

func (s *Schlubs) integrateAndConstrain(centerX, centerY, mobRadius float64) {
	maxDist := mobRadius - SchlubRadius
	maxDistSq := maxDist * maxDist

	for i := range s.schlubs {
		p := &s.schlubs[i]
		p.X += p.VX
		p.Y += p.VY

		// Constrain to mob boundary
		dx := p.X - centerX
		dy := p.Y - centerY
		distSq := dx*dx + dy*dy

		if distSq > maxDistSq {
			dist := math.Sqrt(distSq)
			p.X = centerX + (dx/dist)*maxDist
			p.Y = centerY + (dy/dist)*maxDist

			// Improved velocity handling at boundary
			dot := (p.VX*dx + p.VY*dy) / dist
			if dot > 0 {
				// Reflect velocity component pointing outward
				p.VX -= dot * dx / dist
				p.VY -= dot * dy / dist
				p.VX *= BoundaryDamping
				p.VY *= BoundaryDamping
			}
		}
	}
}

func (s *Schlubs) Draw(screen *ebiten.Image) {
	width := float64(SchlubWidth)
	// scale := SchlubRadius / width

	for _, p := range s.schlubs {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-width, -width)
		// op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(p.X, p.Y)
		color := s.getSchlubColor()
		op.ColorScale.Scale(color[0], color[1], color[2], color[3])

		if p.ID.KindID() == int(world.SchlubKindMonk) {
			screen.DrawImage(s.monkImage, op)
		} else if p.ID.KindID() == int(world.SchlubKindWarrior) {
			screen.DrawImage(s.warriorImage, op)
		} else {
			screen.DrawImage(s.vagrantImage, op)
		}
	}
}

package world

import (
	"math"
)

type ContinentSpec struct {
	Fiefs         int // Number of fiefs per row (e.g., 10 for a 10x10 grid)
	Tiles         int // Number of tiles per fief row(e.g., 10x10)
	TileSize      int // Size of each tile in pixels (e.g., 128)
	FiefPixelSpan int // Size of each fief in pixels, calculated as Tiles * TileSize
	PixelSpan     int // Total pixel span of the continent, calculated as Fiefs * FiefPixelSpan
}

type Continent struct {
	Sneed int64
	Fiefs []*Fief
	Mobs  Mobs
	Fate  Fate
	Specs ContinentSpec
}

func NewContinent(sneed int64) *Continent {
	specs := ContinentSpec{
		Fiefs:    10,  // Default number of fiefs per row
		Tiles:    128, // Default number of tiles per fief row
		TileSize: 64,  // Default size of each tile in pixels
	}
	specs.FiefPixelSpan = specs.Tiles * specs.TileSize  // Size of each fief in pixels
	specs.PixelSpan = specs.Fiefs * specs.FiefPixelSpan // Total pixel span of the continent
	totalFiefs := specs.Fiefs * specs.Fiefs             // Total number of fiefs in the continent

	// Initialize the continent with the given seed and specifications
	fate := NewFate(sneed)
	fiefs := make([]*Fief, totalFiefs)
	for i := range totalFiefs {
		fiefs[i] = NewFief(&fate, i)
	}
	if len(fiefs) == 0 || fiefs[0] == nil {
		panic("failed to create continent: no fiefs generated")
	}

	return &Continent{
		Sneed: sneed,
		Fate:  fate,
		Fiefs: fiefs,
		Specs: specs,
	}
}

// NewMob creates a new Mob instance.
func (c *Continent) NewMob(owner ID, id ID, x, y float64) *Mob {
	mob := &Mob{
		OwnerID: owner,
		ID:      id,
		X:       x,
		Y:       y,
		TargetX: x,
		TargetY: y,
	}
	c.AddMob(mob)
	return mob
}

func (c *Continent) GetFiefAt(x, y int) *Fief {
	// Determine 1-d idx based on x and y coordinates
	if x < 0 || y < 0 || x >= c.Specs.Fiefs || y >= c.Specs.Fiefs {
		// Abso-lute-ly out of bounds
		return nil
	}

	idx := x + y*c.Specs.Fiefs
	if idx < 0 || idx >= len(c.Fiefs) {
		// Fief-ly out of bounds
		return nil
	}
	return c.Fiefs[idx]
}

func (c *Continent) GetContainingFief(x, y float64) *Fief {
	// Translate pixel coordinates to fief grid coordinates
	fiefX := int(math.Floor(x / float64(c.Specs.TileSize)))
	fiefY := int(math.Floor(y / float64(c.Specs.TileSize)))
	return c.GetFiefAt(int(fiefX), int(fiefY))
}

func (c *Continent) GetVisibleFiefs(mob *Mob) []*Fief {
	if mob == nil {
		return nil
	}

	// Slice the fief grid based on the mob's vision radius
	fiefPixelSpan := float64(c.Specs.FiefPixelSpan)
	visionRadius := mob.Vision()
	minX := max(math.Floor((mob.X-visionRadius)/fiefPixelSpan), 0)
	minY := max(math.Floor((mob.Y-visionRadius)/fiefPixelSpan), 0)
	maxX := min(math.Ceil((mob.X+visionRadius)/fiefPixelSpan), float64(c.Specs.Fiefs-1))
	maxY := min(math.Ceil((mob.Y+visionRadius)/fiefPixelSpan), float64(c.Specs.Fiefs-1))

	visibleFiefs := []*Fief{}
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			fief := c.GetFiefAt(int(x), int(y))
			if fief != nil {
				if CircleIntersectsBox(mob.X, mob.Y, mob.Vision(),
					x*fiefPixelSpan, y*fiefPixelSpan,
					fiefPixelSpan, fiefPixelSpan) {
					visibleFiefs = append(visibleFiefs, fief)
				}
			}
		}
	}
	return visibleFiefs
}

func (c *Continent) GetVisibleMobs(mob *Mob) Mobs {
	if mob == nil {
		return nil
	}
	visibleMobs := Mobs{}
	for _, fief := range c.GetVisibleFiefs(mob) {
		if fief == nil {
			continue
		}
		visibleMobs = append(visibleMobs, fief.Mobs.FindVisible(mob.ID)...)
	}
	return visibleMobs
}

func (c *Continent) AddMob(mob *Mob) {
	if mob == nil {
		return
	}

	fief := c.GetContainingFief(mob.X, mob.Y)
	if fief != nil {
		c.Mobs.Add(mob)
		fief.Mobs.Add(mob)
	}
}

func (c *Continent) RemoveMob(mob *Mob) {
	if mob == nil {
		return
	}

	fief := c.GetContainingFief(mob.X, mob.Y)
	if fief != nil {
		fief.Mobs.Remove(mob)
		c.Mobs.Remove(mob)
	}
}

func (c *Continent) MoveMob(mob *Mob, x, y float64) {
	if mob == nil {
		return
	}

	newX := clamp64(x, 0, float64(c.Specs.PixelSpan))
	newY := clamp64(y, 0, float64(c.Specs.PixelSpan))

	currentFief := c.GetContainingFief(mob.X, mob.Y)
	mob.X = newX
	mob.Y = newY

	newFief := c.GetContainingFief(newX, newY)
	if currentFief == newFief {
		// Mob is already in the correct fief, no need to move
		return
	}

	if currentFief != nil {
		currentFief.Mobs.Remove(mob)
	}
	newFief.Mobs.Add(mob)

}

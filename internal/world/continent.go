package world

import (
	"fmt"
	"image/color"
	"math"
)

const ContinientFiefSpan = 35                                 // Number of fiefs per row (e.g., 10 for a 10x10 grid)
const ContinentPixelSpan = ContinientFiefSpan * FiefPixelSpan // Total pixel span of the continent

type Continent struct {
	Sneed uint
	Fiefs []*Fief
	Mobs  Mobs
	Fate  Fate
}

func NewContinent(sneed uint) *Continent {
	// Initialize the continent with the given seed and specifications
	totalFiefs := ContinientFiefSpan * ContinientFiefSpan // Total number of fiefs in the continent
	if totalFiefs <= 0 {
		panic("failed to create continent: total fiefs must be greater than 0")
	}

	fate := NewFate(sneed)
	fiefs := make([]*Fief, totalFiefs)
	for i := range totalFiefs {
		x := i % ContinientFiefSpan
		y := i / ContinientFiefSpan
		if x < 0 || y < 0 || x >= ContinientFiefSpan || y >= ContinientFiefSpan {
			panic("failed to create continent: fief coordinates out of bounds")
		}
		fiefs[i] = NewFief(&fate, x, y)
	}
	if len(fiefs) == 0 || fiefs[0] == nil {
		panic("failed to create continent: no fiefs generated")
	}

	return &Continent{
		Sneed: sneed,
		Fate:  fate,
		Fiefs: fiefs,
	}
}

// NewMob creates a new Mob instance.
func (c *Continent) NewMob(owner ID, id ID, x, y float64) *Mob {
	mob := &Mob{
		OwnerID: owner,
		Color:   color.NRGBA{255, 255, 255, 255}, // Just a default for barbarians.
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
	fiefNum := len(c.Fiefs)

	// Determine 1-d idx based on x and y coordinates
	if x < 0 || y < 0 || x >= fiefNum || y >= fiefNum {
		// Abso-lute-ly out of bounds
		return nil
	}

	idx := x + y*fiefNum
	if idx < 0 || idx >= len(c.Fiefs) {
		// Fief-ly out of bounds
		return nil
	}
	return c.Fiefs[idx]
}

func (c *Continent) GetContainingFief(x, y float64) *Fief {
	// Translate pixel coordinates to fief grid coordinates
	fiefX := int(math.Floor(x / float64(ContinentPixelSpan)))
	fiefY := int(math.Floor(y / float64(ContinentPixelSpan)))
	return c.GetFiefAt(fiefX, fiefY)
}

func (c *Continent) GetVisibleFiefs(mob *Mob) []*Fief {
	if mob == nil {
		return nil
	}

	// Slice the fief grid based on the mob's vision radius
	fiefNum := len(c.Fiefs)
	fiefPixelSpan := float64(FiefPixelSpan)
	visionRadius := mob.Vision()
	minX := max(math.Floor((mob.X-visionRadius)/fiefPixelSpan), 0)
	minY := max(math.Floor((mob.Y-visionRadius)/fiefPixelSpan), 0)
	maxX := min(math.Ceil((mob.X+visionRadius)/fiefPixelSpan), float64(fiefNum-1))
	maxY := min(math.Ceil((mob.Y+visionRadius)/fiefPixelSpan), float64(fiefNum-1))

	visibleFiefs := []*Fief{}
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			fief := c.GetFiefAt(int(x), int(y))
			if fief != nil {
				if CircleIntersectsBox(
					mob.X,
					mob.Y,
					mob.Vision(),
					x*fiefPixelSpan, y*fiefPixelSpan,
					fiefPixelSpan, fiefPixelSpan,
				) {
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
	if fief == nil {
		fmt.Println("Failed to add mob: fief not found for coordinates", mob.X, mob.Y)
		return
	}

	c.Mobs.Add(mob)
	fief.Mobs.Add(mob)
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

	fiefs := c.GetVisibleFiefs(mob)
	for _, fief := range fiefs {
		for mobInFief := range fief.Mobs {
			if mobInFief != mob.ID && mob.Intersects(c.Mobs.FindByID(mobInFief)) {
				// If the mob intersects with another mob in the same fief, do not move
				mob.TargetX = mob.X
				mob.TargetY = mob.Y
				return
			}
		}
	}

	newX := clamp64(x, 0, float64(ContinentPixelSpan))
	newY := clamp64(y, 0, float64(ContinentPixelSpan))
	newFief := c.GetContainingFief(newX, newY)
	if newFief == nil {
		return
	}

	currentFief := c.GetContainingFief(mob.X, mob.Y)
	mob.X = newX
	mob.Y = newY

	if currentFief == newFief {
		// Mob is already in the correct fief, no need to move
		return
	}

	if currentFief != nil {
		currentFief.Mobs.Remove(mob)
	}
	newFief.Mobs.Add(mob)
}

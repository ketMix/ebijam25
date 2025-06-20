package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Fief struct {
	Mobs Mobs
	Img  *ebiten.Image
}

func NewFief(sneed int, size int) *Fief {
	// Determine fief render based on seed.
	img := ebiten.NewImage(size, size)
	randomColor := func() color.Color {
		return color.RGBA{
			R: uint8((sneed*31 + 17) % 256),
			G: uint8((sneed*37 + 29) % 256),
			B: uint8((sneed*41 + 43) % 256),
			A: 50,
		}
	}
	img.Fill(randomColor())
	return &Fief{
		Mobs: Mobs{},
		Img:  img,
	}
}

type Continent struct {
	Sneed    int
	span     int
	FiefSize int
	Fiefs    [][]*Fief
	Mobs     Mobs
}

func NewContinent(sneed int) *Continent {
	span := 10     // Default span
	fiefSize := 64 // Default fief size

	fiefs := make([][]*Fief, span)
	for i := range fiefs {
		fiefs[i] = make([]*Fief, span)
		for j := range fiefs[i] {
			fiefs[i][j] = NewFief(sneed+i+j, fiefSize)
		}
	}
	return &Continent{
		Sneed:    sneed,
		span:     span,
		Fiefs:    fiefs,
		FiefSize: fiefSize,
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

func (c *Continent) isOOB(x, y float64) bool {
	return x < 0 || y < 0 || x >= float64(len(c.Fiefs)*c.FiefSize) || y >= float64(len(c.Fiefs[0])*c.FiefSize)
}

func (c *Continent) GetContainingFief(x, y float64) *Fief {
	if c.isOOB(x, y) {
		return nil // Out of bounds
	}

	fiefX := int(x / float64(c.FiefSize))
	fiefY := int(y / float64(c.FiefSize))
	if fiefX < 0 || fiefY < 0 || fiefX >= len(c.Fiefs) || fiefY >= len(c.Fiefs[0]) {
		return nil // Out of bounds
	}
	return c.Fiefs[fiefX][fiefY]
}

func (c *Continent) GetVisibleFiefs(mob *Mob) []*Fief {
	if mob == nil {
		return nil
	}

	// Slice the fief grid based on the mob's vision radius
	visionRadius := mob.Vision()
	minX := min(int(mob.X-visionRadius)/c.FiefSize, 0)
	minY := min(int(mob.Y-visionRadius)/c.FiefSize, 0)
	maxX := max(int(mob.X+visionRadius)/c.FiefSize, len(c.Fiefs)-1)
	maxY := max(int(mob.Y+visionRadius)/c.FiefSize, len(c.Fiefs[0])-1)
	visibleFiefs := []*Fief{}

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			fief := c.Fiefs[x][y]
			if fief != nil {
				if CircleIntersectsBox(mob.X, mob.Y, mob.Vision(),
					float64(x*c.FiefSize), float64(y*c.FiefSize),
					float64(c.FiefSize), float64(c.FiefSize)) {
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

	maxSize := len(c.Fiefs) * c.FiefSize
	newX := clamp(x, 0, float64(maxSize))
	newY := clamp(y, 0, float64(maxSize))

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

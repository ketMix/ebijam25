package world

import (
	"math"

	"github.com/ketMix/ebijam25/internal/message/event"
)

// Constituent is something participating in a mob.
type Constituent interface {
}

// Mobs is a slice of mobbies, yo.
type Mobs []*Mob

// FindByID searches for a mob by its ID in the Mobs slice.
func (m *Mobs) FindByID(id int) *Mob {
	for _, mob := range *m {
		if mob.ID == id {
			return mob
		}
	}
	return nil
}

// Mob represents a friggin' mob.
type Mob struct {
	ID               int
	X, Y             float64 // Position of the mob in the world
	TargetX, TargetY float64 // Target position to move to
	TargetID         int
	constituents     []Constituent
}

// NewMob creates a new Mob instance.
func NewMob(id int, x, y float64) *Mob {
	return &Mob{
		ID: id,
		X:  x,
		Y:  y,
	}
}

// Update does Mob logic, woo
func (m *Mob) Update( /* some sort of world state */ ) {
	if m.TargetID != 0 {
		// TOOD: Acquire Mob by ID and update TargetX, TargetY.
	}
	if m.X != m.TargetX || m.Y != m.TargetY {
		angleToTarget := math.Atan2(m.TargetY-m.Y, m.TargetX-m.X)
		dx := math.Cos(angleToTarget)
		dy := math.Sin(angleToTarget)
		x := m.X + dx
		y := m.Y + dy
		if math.Abs(x-m.TargetX) < 1 && math.Abs(y-m.TargetY) < 1 {
			x = m.TargetX
			y = m.TargetY
		}
		EventBus.Publish(&event.MobPosition{ID: m.ID, X: int(x), Y: int(y)})
	}
}

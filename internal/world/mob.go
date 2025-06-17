package world

import (
	"math"
	"slices"

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

// Add appends a new mob to the Mobs slice.
func (m *Mobs) Add(mob *Mob) {
	if slices.Contains(*m, mob) {
		return
	}
	*m = append(*m, mob)
}

// Remove deletes a mob from the Mobs slice.
func (m *Mobs) Remove(mob *Mob) {
	for i, existingMob := range *m {
		if existingMob == mob {
			*m = append((*m)[:i], (*m)[i+1:]...)
			return
		}
	}
}

// Mob represents a friggin' mob.
type Mob struct {
	OwnerID          int // ID of the owner(player)
	ID               int
	X, Y             float64 // Position of the mob in the world
	TargetX, TargetY float64 // Target position to move to
	TargetID         int
	constituents     []Constituent
}

// NewMob creates a new Mob instance.
func NewMob(owner int, id int, x, y float64) *Mob {
	return &Mob{
		ID: id,
		X:  x,
		Y:  y,
	}
}

// Update does Mob logic, woo
func (m *Mob) Update(state *State) {
	// Acquire our target mob if we have one set.
	if m.TargetID != 0 {
		if mob := state.Mobs.FindByID(m.TargetID); mob != nil {
			m.TargetX = mob.X
			m.TargetY = mob.Y
		} else {
			m.TargetID = 0 // Reset if target mob is not found
		}
	}

	// Move towards our destiny.
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
		state.EventBus.Publish(&event.MobPosition{ID: m.ID, X: int(x), Y: int(y)})
	}
}

package world

import (
	"image/color"
	"math"
	"slices"

	"github.com/ketMix/ebijam25/internal/message/event"
)

const (
	FloatScale       float64 = 7.0
	MaxSchlubsPerMob int     = 1000 // Max number of schlubs in a mob
)

// Mobs is a slice of mobbies, yo.
type Mobs []*Mob

// FindVisible returns a slice of mobs that are visible to the mob with the given ID.
func (m *Mobs) FindVisible(mobID ID) Mobs {
	sourceMob := m.FindByID(mobID)
	if sourceMob == nil {
		return nil // Mob not found, return empty slice
	}
	var visibleMobs Mobs
	for _, mob := range *m {
		if CircleIntersectsCircle(sourceMob.X, sourceMob.Y, sourceMob.Vision(), mob.X, mob.Y, mob.Radius()) {
			visibleMobs = append(visibleMobs, mob)
		}
	}
	return visibleMobs
}

// FindByID searches for a mob by its ID in the Mobs slice.
func (m *Mobs) FindByID(id ID) *Mob {
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
	OwnerID          ID // ID of the owner(player)
	ID               ID
	Color            color.NRGBA // Color of the mob in NRGBA format, this is duplicated from the owning player.
	X, Y             float64     // Position of the mob in the world
	TargetX, TargetY float64     // Target position to move to
	TargetID         ID
	Stats            *Stats // Stats of the mob
	Schlubs          []SchlubID
	OuterKind        SchlubID // Outer kind of the mob, used for formation
}

// Update does Mob logic, woo
func (m *Mob) Update(state *State) {
	speed := 0.7 * float64(state.Tickrate)

	// Acquire our target mob if we have one set.
	if m.TargetID != 0 {
		if mob := state.Continent.Mobs.FindByID(m.TargetID); mob != nil {
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
		x := m.X + dx*speed
		y := m.Y + dy*speed

		if math.Abs(x-m.TargetX) < speed {
			x = m.TargetX
		}
		if math.Abs(y-m.TargetY) < speed {
			y = m.TargetY
		}

		intX := int(x * FloatScale)
		intY := int(y * FloatScale)
		state.EventBus.Publish(&event.MobPosition{ID: m.ID, X: intX, Y: intY})
	}
}

func (m *Mob) AddSchlub(schlub ...SchlubID) {
	m.Schlubs = append(m.Schlubs, schlub...)
}

func (m *Mob) RemoveSchlub(schlub ...SchlubID) {
	for _, id := range schlub {
		for i, existingSchlub := range m.Schlubs {
			if existingSchlub == id {
				m.Schlubs = append(m.Schlubs[:i], m.Schlubs[i+1:]...)
				break // Exit the loop after removing the schlub
			}
		}
	}
}

// Radius calculates the radius of the mob based on the number of constituents.
func (m *Mob) Radius() float64 {
	if len(m.Schlubs) == 0 {
		return 1
	}
	return float64(len(m.Schlubs)) * 0.1
}

func (m *Mob) Spread() float64 {
	if len(m.Schlubs) == 0 {
		return 1
	}
	return float64(len(m.Schlubs)) * 2
}

// Vision returns the mob's vision radius.
func (m *Mob) Vision() float64 {
	vision := math.Max(200, math.Log(m.Radius())*100)
	return vision
}

// Intersects checks if the mob's circle edge intersects with another mob's circle edge.
func (m *Mob) Intersects(other *Mob) bool {
	if other == nil {
		return false
	}
	return CircleIntersectsCircle(m.X, m.Y, m.Radius(), other.X, other.Y, other.Radius())
}

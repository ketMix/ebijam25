package server

import (
	"github.com/ketMix/ebijam25/internal/world"
)

// UpdateMob updates the state of a mob in the game.
func (g *Game) UpdateMob(mob *world.Mob) {
	mob.Update(&g.State)
}

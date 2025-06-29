package client

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ketMix/ebijam25/internal/world"
)

// DrawMob draws a mobs on the screen.
func (g *Game) DrawMob(screen *ebiten.Image, mob *world.Mob, simple bool) {
	if mob == nil {
		return
	}

	// Get particle system for this mob
	ps, exists := g.schlubSystem[mob.ID]
	if !simple && exists && ps != nil {
		// Draw particles with fluid effect
		ps.Draw(screen, !simple) // just for now
	} else {
		// Fallback to simple circle if no particle system
		radius := mob.CombatRadius()
		vector.DrawFilledCircle(screen, float32(mob.X), float32(mob.Y), float32(radius), mob.Color, true)
		// Draw schlub count
		countText := fmt.Sprintf("%d", len(mob.Schlubs))
		ebitenutil.DebugPrintAt(screen, countText, int(mob.X)-10, int(mob.Y)-5)
	}

	// Draw vision circle for local player
	if mob.ID == g.MobID {
		vision := mob.Vision()
		vector.StrokeCircle(screen, float32(mob.X), float32(mob.Y), float32(vision), 4, color.NRGBA{0, 0, 255, 64}, false)
	}

	// Also draw a "combat" circle for any mob.
	vector.StrokeCircle(screen, float32(mob.X), float32(mob.Y), float32(mob.CombatRadius()), 4, color.NRGBA{255, 0, 0, 128}, false)
}

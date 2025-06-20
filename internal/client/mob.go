package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ketMix/ebijam25/internal/world"
)

// DrawMob draws a mobs on the screen.
func (g *Game) DrawMob(screen *ebiten.Image, mob *world.Mob) {
	if mob == nil {
		return
	}
	radius := mob.Radius()
	vector.StrokeCircle(screen, float32(mob.X), float32(mob.Y), float32(radius), 1, color.NRGBA{255, 0, 255, 255}, true)

	// Draw vision circle for local player.
	if mob.ID == g.MobID {
		vision := mob.Vision()
		vector.StrokeCircle(screen, float32(mob.X), float32(mob.Y), float32(vision), 1, color.NRGBA{0, 255, 0, 64}, false)
	}
}

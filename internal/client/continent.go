package client

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (g *Game) DrawContinent(screen *ebiten.Image) {
	if g.Continent == nil || g.Continent.Fiefs == nil {
		return
	}
	ops := &ebiten.DrawImageOptions{}

	for x, row := range g.Continent.Fiefs {
		for y, fief := range row {
			if fief == nil {
				continue
			}

			ops.GeoM.Reset()
			ops.GeoM.Translate(float64(x*g.Continent.FiefSize), float64(y*g.Continent.FiefSize))
			if g.Debug {
				borderSize := float32(g.Continent.FiefSize / 10)
				vector.DrawFilledRect(screen,
					float32(x*g.Continent.FiefSize),
					float32(y*g.Continent.FiefSize),
					float32(g.Continent.FiefSize),
					float32(g.Continent.FiefSize),
					color.White,
					true,
				)
				vector.DrawFilledRect(screen,
					float32(x*g.Continent.FiefSize)-borderSize,
					float32(y*g.Continent.FiefSize)-borderSize,
					float32(g.Continent.FiefSize)+2*borderSize,
					float32(g.Continent.FiefSize)+2*borderSize,
					color.Gray{},
					true,
				)

				ebitenutil.DebugPrintAt(screen,
					fmt.Sprintf("(%d,%d)\nMobs: %d", x, y, len(fief.Mobs)),
					x*g.Continent.FiefSize+2, y*g.Continent.FiefSize+2)
			} else if img := g.fiefImages[y][x]; img != nil {
				// We do be assuming.
				screen.DrawImage(img, ops)
			}

			for _, mob := range fief.Mobs {
				g.DrawMob(screen, mob)
			}
		}
	}
}

package client

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/ketMix/ebijam25/internal/world"
)

func DrawTile(screen *ebiten.Image, t world.Terrain, x, y, size float32) {
	if screen == nil || t == world.TerrainNone {
		return
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	opts.GeoM.Scale(float64(size)/float64(world.TileSize), float64(size)/float64(world.TileSize))
	opts.ColorScale.ScaleAlpha(0.85)
	DrawTerrain(screen, opts, t)
}

func (g *Game) DrawFiefs(screen *ebiten.Image) {
	if screen == nil {
		g.log.Error("nil screen provided for drawing fiefs")
		return
	}
	fiefs := g.Continent.Fiefs
	if fiefs == nil || len(fiefs) == 0 {
		g.log.Error("no fiefs to draw")
		return
	}

	// Create fief images if not already created
	if len(fiefImages) == 0 {
		fiefNum := len(fiefs)
		fiefImages = make([]*ebiten.Image, fiefNum)
		tileSize := float32(world.TileSize)

		for i, fief := range fiefs {
			img := ebiten.NewImage(world.FiefPixelSpan, world.FiefPixelSpan)
			fiefImages[i] = img
			for j := range world.FiefSize {
				for k := range world.FiefSize {
					idx := j + k*world.FiefSize
					if idx >= len(fief.Tiles) {
						g.log.Warn(fmt.Sprintf("index out of bounds for fief %d at (%d, %d)", i, j, k))
						continue
					}
					tile := fief.Tiles[idx]
					tileX := float32(j * world.TileSize)
					tileY := float32(k * world.TileSize)
					DrawTile(img, tile.Terrain, tileX, tileY, tileSize)
				}
			}
		}
	}
	if g.Debug {
		for idx, fief := range fiefs {
			if fief == nil {
				continue
			}
			x := (idx % world.ContinientFiefSpan) * world.FiefPixelSpan
			y := (idx / world.ContinientFiefSpan) * world.FiefPixelSpan
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Fief %d", idx), x+2, y+2)
		}
	} else {
		for idx, fief := range fiefs {
			if fief == nil || fiefImages[idx] == nil {
				g.log.Warn("fief or its image is nil", "index", idx)
				continue
			}
			ops := &ebiten.DrawImageOptions{}
			ops.GeoM.Translate(float64(fief.X), float64(fief.Y))
			screen.DrawImage(fiefImages[idx], ops)
		}
	}

	for _, mob := range g.Continent.Mobs {
		if mob == nil {
			g.log.Warn("nil mob found in continent mobs")
			continue
		}
		g.DrawMob(screen, mob)
	}

}

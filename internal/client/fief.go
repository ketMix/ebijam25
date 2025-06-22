package client

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ketMix/ebijam25/internal/world"
)

func DrawTile(screen *ebiten.Image, t world.Terrain, x, y, size float32) {
	if t == world.TerrainNone {
		return
	}

	var a uint8 = 100
	c := color.NRGBA{128, 128, 128, 255} // Default color for unknown terrain
	switch t {
	case world.TerrainGrass:
		c = color.NRGBA{34, 139, 34, 255}
	case world.TerrainWater:
		c = color.NRGBA{0, 0, 255, 255}
	case world.TerrainMountain:
		c = color.NRGBA{139, 137, 137, 255}
	case world.TerrainForest:
		c = color.NRGBA{34, 139, 34, 255}
	case world.TerrainDesert:
		c = color.NRGBA{210, 180, 140, 255}
	case world.TerrainSwamp:
		c = color.NRGBA{85, 107, 47, 255}
	case world.TerrainSnow:
		c = color.NRGBA{255, 250, 250, 255}
	default:
		break
	}
	c.A = a
	vector.DrawFilledRect(screen, x, y, size, size, c, true)
}

// func (g *Game) MakeFiefImages(sneed int64) {
// 	fiefNum := world.ContinientFiefSpan * world.ContinientFiefSpan
// 	fiefs := make([]*ebiten.Image, fiefNum)

// 	for i := range fiefNum {
// 		img := ebiten.NewImage(world.FiefSize, world.FiefSize)
// 		fsneed := int(sneed) + i
// 		randomColor := func() color.Color {
// 			return color.NRGBA{
// 				R: uint8((fsneed*31 + 17) % 256),
// 				G: uint8((fsneed*37 + 29) % 256),
// 				B: uint8((fsneed*41 + 43) % 256),
// 				A: 50,
// 			}
// 		}
// 		img.Fill(randomColor())
// 		fiefs[i] = img
// 	}
// 	g.fiefImages = fiefs
// }

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
			for _, mob := range fief.Mobs {
				g.DrawMob(screen, mob)
			}
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
			for _, mob := range fief.Mobs {
				g.DrawMob(screen, mob)
			}
		}
	}

}

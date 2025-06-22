package client

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/ketMix/ebijam25/internal/world"
)

var continentImage *ebiten.Image

func DrawTile(screen *ebiten.Image, t world.Terrain, x, y, size float32) {
	if t == world.TerrainNone {
		return
	}

	c := color.RGBA{128, 128, 128, 255} // Default color for unknown terrain
	switch t {
	case world.TerrainGrass:
		c = color.RGBA{34, 139, 34, 255}
	case world.TerrainWater:
		c = color.RGBA{0, 0, 255, 150}
	case world.TerrainMountain:
		c = color.RGBA{139, 137, 137, 255}
	case world.TerrainForest:
		c = color.RGBA{34, 139, 34, 255}
	case world.TerrainDesert:
		c = color.RGBA{210, 180, 140, 255}
	case world.TerrainSwamp:
		c = color.RGBA{85, 107, 47, 255}
	case world.TerrainSnow:
		c = color.RGBA{255, 250, 250, 255}
	default:
		break
	}
	fmt.Println("Drawing tile at", x, y, "with size", size, "and color", c)
	vector.DrawFilledRect(screen, x, y, size, size, c, true)
}

func (g *Game) DrawFiefs(screen *ebiten.Image) {
	fiefs := g.Continent.Fiefs
	if fiefs == nil || len(fiefs) == 0 {
		g.log.Error("no fiefs to draw")
		return
	}

	for _, fief := range fiefs {
		if fief == nil {
			continue
		}

		fiefMinX := fief.Index % g.Continent.Specs.Fiefs
		fiefMinY := fief.Index / g.Continent.Specs.Fiefs
		tileSize := float32(fief.TileSpan)
		for x := range fief.Span {
			for y := range fief.Span {
				tile := fief.GetTileAt(x, y)
				if tile != nil {
					tileX := float32(fiefMinX*fief.TileSpan + x*fief.TileSpan)
					tileY := float32(fiefMinY*fief.TileSpan + y*fief.TileSpan)
					DrawTile(screen, tile.Terrain, tileX, tileY, tileSize)
				}
			}
		}
	}

}
func (g *Game) DrawContinent(screen *ebiten.Image) {
	if g.Continent == nil || g.Continent.Fiefs == nil {
		return
	}
	if continentImage == nil {
		pix := int(g.Continent.Specs.PixelSpan)
		continentImage = ebiten.NewImage(pix, pix)
		g.DrawFiefs(continentImage)
	}
	ops := &ebiten.DrawImageOptions{}
	screen.DrawImage(continentImage, ops)
	for _, mob := range g.Continent.Mobs {
		g.DrawMob(screen, mob)
	}
}

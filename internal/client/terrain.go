package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
)

var terrainCache map[world.Terrain]*ebiten.Image

func DrawTerrain(screen *ebiten.Image, opts *ebiten.DrawImageOptions, terrain world.Terrain) {
	if terrainCache == nil {
		terrainCache = make(map[world.Terrain]*ebiten.Image)
	}
	img, ok := terrainCache[terrain]
	if !ok {

		// TODO: terrain texs
		var a uint8 = 100
		c := color.NRGBA{128, 128, 128, 255}
		switch terrain {
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
		img = ebiten.NewImage(world.TileSize, world.TileSize)
		img.Fill(c)
		terrainCache[terrain] = img
	}

	screen.DrawImage(img, opts)
}

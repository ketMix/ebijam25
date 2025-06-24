package client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
	"github.com/ketMix/ebijam25/stuff"
)

var terrainCache []*ebiten.Image

func DrawTerrain(screen *ebiten.Image, opts *ebiten.DrawImageOptions, terrain world.Terrain) {
	if len(terrainCache) == 0 {
		terrainCache = make([]*ebiten.Image, world.TerrainCount)
	}
	if terrain <= world.TerrainNone || int(terrain) >= int(world.TerrainCount) {
		screen.Fill(color.NRGBA{255, 0, 0, 255}) // Invalid terrain
		return
	}
	img := terrainCache[terrain]
	if img == nil {
		img = stuff.GetImage(terrain.ImageName())
		if img == nil {
			img = ebiten.NewImage(world.TileSize, world.TileSize)
			img.Fill(color.White) // Fallback to white if image not found
		}
		terrainCache[terrain] = img
	}
	opts.ColorScale.ScaleAlpha(0.85)
	screen.DrawImage(img, opts)
}

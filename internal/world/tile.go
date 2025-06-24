package world

const TileSize = 16 // Size of each tile in pixels

type Tile struct {
	Terrain Terrain
}

func NewTile(fate *Fate, x, y float64) Tile {
	return Tile{
		Terrain: NewTerrain(fate, x, y),
	}
}

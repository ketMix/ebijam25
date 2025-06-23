package world

const FiefSize = 32                       // Number of tiles per fief row (e.g., 64x64)
const FiefTiles = FiefSize * FiefSize     // Total number of tiles in a fief
const FiefPixelSpan = FiefSize * TileSize // Total pixel span of a fief

type Fief struct {
	X, Y      float64
	Name      string
	Mobs      Mobs
	Tiles     []Tile
	modifiers []Modifier
}

func NewFief(fate *Fate, x, y int) *Fief {
	tiles := make([]Tile, FiefTiles)
	fiefX := float64(x * FiefPixelSpan)
	fiefY := float64(y * FiefPixelSpan)
	for i := range tiles {
		tX := i % FiefSize
		tY := i / FiefSize
		tileX := float64(tX*TileSize) + fiefX
		tileY := float64(tY*TileSize) + fiefY + float64(TileSize)
		tiles[i] = NewTile(fate, tileX, tileY)
	}

	return &Fief{
		X:         fiefX,
		Y:         fiefY,
		Name:      "Fief",
		Mobs:      Mobs{},
		Tiles:     tiles,
		modifiers: []Modifier{},
	}
}

func (f *Fief) GetTileAt(x, y float64) *Tile {
	if f == nil || len(f.Tiles) == 0 {
		return nil
	}
	if x < f.X || y < f.Y {
		return nil
	}
	if x >= f.X+float64(FiefPixelSpan) || y >= f.Y+float64(FiefPixelSpan) {
		return nil
	}

	tileX := int((x - f.X) / TileSize)
	tileY := int((y - f.Y) / TileSize)
	if tileX < 0 || tileY < 0 || tileX >= FiefSize || tileY >= FiefSize {
		return nil
	}
	tileIndex := tileX + tileY*FiefSize
	if tileIndex < 0 || tileIndex >= len(f.Tiles) {
		return nil
	}
	return &f.Tiles[tileIndex]
}

func (f *Fief) GetModifiers(mob *Mob) []Modifier {
	if mob == nil {
		return f.modifiers
	}
	tile := f.GetTileAt(mob.X, mob.Y)
	if tile != nil {
		modifiers := make([]Modifier, 0, len(f.modifiers)+1)
		copy(modifiers, f.modifiers)
		return append(modifiers, []Modifier{tile.Terrain.GetModifier()}...)
	}
	return f.modifiers
}

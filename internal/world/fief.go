package world

type Tile struct {
	Terrain Terrain
}

type Fief struct {
	Span      int // Size of the fief in tiles (e.g., 10x10)
	Index     int
	Name      string
	Mobs      Mobs
	Tiles     []Tile
	TileSpan  int // Span of each tile in pixels, can be used for rendering
	modifiers []Modifier
}

func NewFief(fate *Fate, idx int) *Fief {
	span := 10      // Default span for a fief, can be adjusted as needed
	tileSpan := 128 // Default tile span in pixels, can be adjusted as needed
	fiefSeed := fate.Eval64(float64(idx))
	fiefFate := NewFate(int64(fiefSeed))

	tiles := make([]Tile, span*span)

	for i := range span {
		for j := range span {
			idx := i + j*span
			tiles[idx] = Tile{
				Terrain: NewTerrain(int(fiefFate.Determine(float64(i), float64(j)))),
			}
		}
	}

	return &Fief{
		Name:      "Fief",
		Span:      span,
		Index:     idx,
		Mobs:      Mobs{},
		Tiles:     tiles,
		TileSpan:  tileSpan,
		modifiers: []Modifier{},
	}
}

func (f *Fief) GetTileAt(x, y int) *Tile {
	if x < 0 || y < 0 || x >= len(f.Tiles) || y >= len(f.Tiles) {
		return nil // Out of bounds
	}
	idx := x + y*len(f.Tiles)
	if idx < 0 || idx >= len(f.Tiles) {
		return nil // Out of bounds
	}
	return &f.Tiles[idx]
}

func (f *Fief) GetModifiers(mob *Mob) []Modifier {
	if mob == nil {
		return f.modifiers
	}

	tile := f.GetTileAt(int(mob.X), int(mob.Y))
	if tile != nil {
		modifiers := make([]Modifier, 0, len(f.modifiers)+1)
		copy(modifiers, f.modifiers)
		return append(modifiers, []Modifier{tile.Terrain.GetModifier()}...)
	}
	return f.modifiers
}

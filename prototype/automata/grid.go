package main

// Cell represents a single cell in the automata grid
type Cell struct {
	Bullet        *Bullet
	X, Y          int
	deathsByChain map[int64]DeathInfo // Track deaths by chain ID
}

// DeathInfo tracks when a bullet died and its cooldown
type DeathInfo struct {
	timestamp int64 // When the bullet died
	cooldown  int   // How many steps to wait
}

// NewCell creates a new cell
func NewCell(x, y int) *Cell {
	return &Cell{
		X:             x,
		Y:             y,
		deathsByChain: make(map[int64]DeathInfo),
	}
}

// IsAlive returns true if the cell has a living bullet
func (c *Cell) IsAlive() bool {
	return c.Bullet != nil && c.Bullet.IsAlive()
}

// GetState returns the current state of the cell
func (c *Cell) GetState() CellState {
	if c.IsAlive() {
		return Alive
	}
	return Empty
}

// SetBullet sets the bullet in this cell
func (c *Cell) SetBullet(bullet *Bullet) {
	c.Bullet = bullet
	if bullet != nil {
		bullet.X = c.X
		bullet.Y = c.Y
	}
}

// RemoveBullet removes the bullet from this cell
func (c *Cell) RemoveBullet() *Bullet {
	bullet := c.Bullet
	c.Bullet = nil
	return bullet
}

// MarkDeath marks this cell as having a recent death for a specific chain
func (c *Cell) MarkDeath(currentStep int64, cooldownDuration int, chainID int64) {
	c.deathsByChain[chainID] = DeathInfo{
		timestamp: currentStep,
		cooldown:  cooldownDuration,
	}
}

// IsInDeathCooldown returns true if this cell is still in death cooldown for a specific chain
func (c *Cell) IsInDeathCooldown(currentStep int64, chainID int64) bool {
	deathInfo, exists := c.deathsByChain[chainID]
	if !exists || deathInfo.cooldown <= 0 {
		return false
	}
	return (currentStep - deathInfo.timestamp) < int64(deathInfo.cooldown)
}

// HasAnyCooldown returns true if any chain has a death cooldown in this cell
func (c *Cell) HasAnyCooldown(currentStep int64) bool {
	for chainID := range c.deathsByChain {
		if c.IsInDeathCooldown(currentStep, chainID) {
			return true
		}
	}
	return false
}

// CanAcceptNewBullet returns true if a new bullet can be placed in this cell
func (c *Cell) CanAcceptNewBullet(currentStep int64, chainID int64) bool {
	// Can't place if there's already a living bullet
	if c.IsAlive() {
		return false
	}

	// Can't place if in death cooldown for this chain
	if c.IsInDeathCooldown(currentStep, chainID) {
		return false
	}

	return true
}

// Grid represents the 2D cellular automata grid
type Grid struct {
	width, height int
	cells         [][]*Cell
}

// NewGrid creates a new grid with the specified dimensions
func NewGrid(width, height int) *Grid {
	cells := make([][]*Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]*Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = NewCell(x, y)
		}
	}

	return &Grid{
		width:  width,
		height: height,
		cells:  cells,
	}
}

// GetWidth returns the grid width
func (g *Grid) GetWidth() int {
	return g.width
}

// GetHeight returns the grid height
func (g *Grid) GetHeight() int {
	return g.height
}

// GetCell returns the cell at the given coordinates (thread-safe)
func (g *Grid) GetCell(x, y int) *Cell {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return nil
	}
	return g.cells[y][x]
}

// SetBulletAt places a bullet at the given coordinates (thread-safe)
func (g *Grid) SetBulletAt(x, y int, bullet *Bullet) bool {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return false
	}

	g.cells[y][x].SetBullet(bullet)
	return true
}

// RemoveBulletAt removes the bullet at the given coordinates (thread-safe)
func (g *Grid) RemoveBulletAt(x, y int) *Bullet {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return nil
	}
	return g.cells[y][x].RemoveBullet() // Remove bullet
}

// GetArea3x3 returns a 3x3 area of cell states centered at (x, y)
func (g *Grid) GetArea3x3(x, y int) [3][3]CellState {
	var area [3][3]CellState

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && ny >= 0 && nx < g.width && ny < g.height {
				area[dy+1][dx+1] = g.cells[ny][nx].GetState()
			} else {
				area[dy+1][dx+1] = Empty // Out of bounds is considered empty
			}
		}
	}

	return area
}

// IsInBounds checks if the given coordinates are within the grid
func (g *Grid) IsInBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < g.width && y < g.height
}

// ForEachCell iterates over all cells in the grid (thread-safe read)
func (g *Grid) ForEachCell(fn func(x, y int, cell *Cell)) {
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			fn(x, y, g.cells[y][x])
		}
	}
}

// GetLivingBullets returns all living bullets in the grid
func (g *Grid) GetLivingBullets() []*Bullet {
	var bullets []*Bullet
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			if g.cells[y][x].IsAlive() {
				bullets = append(bullets, g.cells[y][x].Bullet)
			}
		}
	}
	return bullets
}

// Clear removes all bullets from the grid
func (g *Grid) Clear() {
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			g.cells[y][x].RemoveBullet()
		}
	}
}

package main

import (
	"sync"
	"sync/atomic"
)

// ObjectPool provides high-performance object pooling for frequently allocated objects
type ObjectPool struct {
	bullets  []*Bullet
	actions  []Action
	rulesets []*RuleSet
	maxSize  int
}

// NewObjectPool creates a new object pool
func NewObjectPool(maxSize int) *ObjectPool {
	return &ObjectPool{
		bullets:  make([]*Bullet, 0, maxSize),
		actions:  make([]Action, 0, maxSize*4), // Actions might be more frequent
		rulesets: make([]*RuleSet, 0, maxSize/10),
		maxSize:  maxSize,
	}
}

// GetBullet gets a bullet from the pool or creates a new one
func (p *ObjectPool) GetBullet() *Bullet {
	if len(p.bullets) > 0 {
		bullet := p.bullets[len(p.bullets)-1]
		p.bullets = p.bullets[:len(p.bullets)-1]
		return bullet
	}

	// Create new bullet if pool is empty
	return &Bullet{}
}

// PutBullet returns a bullet to the pool
func (p *ObjectPool) PutBullet(bullet *Bullet) {
	if bullet == nil {
		return
	}

	// Reset bullet state properly
	bullet.X = 0
	bullet.Y = 0
	bullet.ruleSet = nil
	bullet.generation = 0
	bullet.colorR = 255
	bullet.colorG = 255
	bullet.colorB = 255

	// Reset atomic fields properly
	atomic.StoreInt32(&bullet.alive, 0)
	atomic.StoreInt32(&bullet.lifetime, 0)
	atomic.StoreInt32(&bullet.maxLifetime, 100)
	atomic.StoreInt32(&bullet.displayFrames, 0)

	if len(p.bullets) < p.maxSize {
		p.bullets = append(p.bullets, bullet)
	}
}

// GetActionSlice gets a slice of actions from the pool
func (p *ObjectPool) GetActionSlice() []Action {
	if len(p.actions) >= 16 {
		slice := p.actions[:16]
		p.actions = p.actions[16:]
		return slice[:0] // Return with zero length but capacity
	}

	return make([]Action, 0, 16)
}

// PutActionSlice returns an action slice to the pool
func (p *ObjectPool) PutActionSlice(actions []Action) {
	if actions == nil || cap(actions) < 16 {
		return
	}

	// Clear the slice but keep capacity
	actions = actions[:cap(actions)]
	for i := range actions {
		actions[i] = Action{} // Clear action data
	}

	if len(p.actions)+len(actions) <= p.maxSize*4 {
		p.actions = append(p.actions, actions...)
	}
}

// SpatialIndex provides spatial partitioning for efficient collision detection
type SpatialIndex struct {
	gridSize   int
	cellSize   float64
	gridWidth  int
	gridHeight int
	cells      [][]*SpatialCell
}

// SpatialCell represents a cell in the spatial index
type SpatialCell struct {
	bullets []*Bullet
	dirty   bool
}

// NewSpatialIndex creates a new spatial index
func NewSpatialIndex(worldWidth, worldHeight, cellSize int) *SpatialIndex {
	gridWidth := (worldWidth + cellSize - 1) / cellSize
	gridHeight := (worldHeight + cellSize - 1) / cellSize

	cells := make([][]*SpatialCell, gridHeight)
	for y := 0; y < gridHeight; y++ {
		cells[y] = make([]*SpatialCell, gridWidth)
		for x := 0; x < gridWidth; x++ {
			cells[y][x] = &SpatialCell{
				bullets: make([]*Bullet, 0, 8), // Pre-allocate for typical case
			}
		}
	}

	return &SpatialIndex{
		cellSize:   float64(cellSize),
		gridWidth:  gridWidth,
		gridHeight: gridHeight,
		cells:      cells,
	}
}

// Clear clears all bullets from the spatial index
func (si *SpatialIndex) Clear() {
	for y := 0; y < si.gridHeight; y++ {
		for x := 0; x < si.gridWidth; x++ {
			cell := si.cells[y][x]
			cell.bullets = cell.bullets[:0] // Clear but keep capacity
			cell.dirty = false
		}
	}
}

// Insert adds a bullet to the spatial index
func (si *SpatialIndex) Insert(bullet *Bullet) {
	if bullet == nil {
		return
	}

	gridX := int(float64(bullet.X) / si.cellSize)
	gridY := int(float64(bullet.Y) / si.cellSize)

	if gridX < 0 || gridY < 0 || gridX >= si.gridWidth || gridY >= si.gridHeight {
		return
	}

	cell := si.cells[gridY][gridX]
	cell.bullets = append(cell.bullets, bullet)
	cell.dirty = true
}

// QueryArea returns all bullets in the given area
func (si *SpatialIndex) QueryArea(x, y, radius int) []*Bullet {
	var result []*Bullet

	// Calculate grid bounds
	minGridX := int(float64(x-radius) / si.cellSize)
	maxGridX := int(float64(x+radius) / si.cellSize)
	minGridY := int(float64(y-radius) / si.cellSize)
	maxGridY := int(float64(y+radius) / si.cellSize)

	// Clamp to grid bounds
	if minGridX < 0 {
		minGridX = 0
	}
	if maxGridX >= si.gridWidth {
		maxGridX = si.gridWidth - 1
	}
	if minGridY < 0 {
		minGridY = 0
	}
	if maxGridY >= si.gridHeight {
		maxGridY = si.gridHeight - 1
	}

	// Collect bullets from relevant cells
	for gy := minGridY; gy <= maxGridY; gy++ {
		for gx := minGridX; gx <= maxGridX; gx++ {
			cell := si.cells[gy][gx]
			for _, bullet := range cell.bullets {
				if bullet.IsAlive() {
					// Check if bullet is actually within radius
					dx := bullet.X - x
					dy := bullet.Y - y
					if dx*dx+dy*dy <= radius*radius {
						result = append(result, bullet)
					}
				}
			}
		}
	}

	return result
}

// BulletManager manages bullet lifecycle with high performance
type BulletManager struct {
	pool         *ObjectPool
	spatialIndex *SpatialIndex
	liveBullets  []*Bullet
	mutex        sync.RWMutex
}

// NewBulletManager creates a new bullet manager
func NewBulletManager(worldWidth, worldHeight int) *BulletManager {
	return &BulletManager{
		pool:         NewObjectPool(1000),
		spatialIndex: NewSpatialIndex(worldWidth, worldHeight, 10),
		liveBullets:  make([]*Bullet, 0, 500),
	}
}

// CreateBullet creates a new bullet using object pooling
func (bm *BulletManager) CreateBullet(x, y int, ruleSet *RuleSet) *Bullet {
	bullet := bm.pool.GetBullet()

	// Properly initialize bullet from pool
	bm.initializeBullet(bullet, x, y, ruleSet)
	bm.liveBullets = append(bm.liveBullets, bullet)

	// Add to spatial index
	bm.spatialIndex.Insert(bullet)

	return bullet
}

// initializeBullet properly initializes a bullet from the pool
func (bm *BulletManager) initializeBullet(bullet *Bullet, x, y int, ruleSet *RuleSet) {
	// Always clone the rule set so bullets maintain their behavior independently
	var clonedRuleSet *RuleSet
	if ruleSet != nil {
		clonedRuleSet = ruleSet.Clone()
	}

	// Initialize all fields properly
	bullet.X = x
	bullet.Y = y
	bullet.ruleSet = clonedRuleSet
	bullet.generation = 0
	bullet.id = atomic.AddInt64(&bulletIDCounter, 1)
	bullet.colorR = 255
	bullet.colorG = 255
	bullet.colorB = 255

	// Reset atomic fields properly
	atomic.StoreInt32(&bullet.alive, 1)
	atomic.StoreInt32(&bullet.lifetime, 0)
	atomic.StoreInt32(&bullet.displayFrames, 0)

	// Set default lifetime
	if clonedRuleSet != nil {
		atomic.StoreInt32(&bullet.maxLifetime, int32(clonedRuleSet.GetDefaultLifetime()))
	} else {
		atomic.StoreInt32(&bullet.maxLifetime, 100)
	}
}

// Update updates all bullets and performs cleanup
func (bm *BulletManager) Update() {
	// Clear spatial index for this frame
	bm.spatialIndex.Clear()

	// Update bullets and separate living from dead in-place
	writeIndex := 0
	for _, bullet := range bm.liveBullets {
		if bullet.IsAlive() {
			bullet.Age()

			if bullet.IsAlive() {
				// Keep this bullet - move it to the write position
				bm.liveBullets[writeIndex] = bullet
				writeIndex++
				bm.spatialIndex.Insert(bullet)
			} else {
				// Bullet died during aging - return to pool
				bm.pool.PutBullet(bullet)
			}
		} else {
			// Bullet was already dead - return to pool
			bm.pool.PutBullet(bullet)
		}
	}

	// Truncate the slice to remove dead bullets
	bm.liveBullets = bm.liveBullets[:writeIndex]
}

// RemoveBullet immediately removes a bullet from the manager
func (bm *BulletManager) RemoveBullet(bulletToRemove *Bullet) {
	if bulletToRemove == nil {
		return
	}

	// Remove from live bullets slice
	for i, bullet := range bm.liveBullets {
		if bullet == bulletToRemove {
			// Remove by swapping with last element and truncating
			bm.liveBullets[i] = bm.liveBullets[len(bm.liveBullets)-1]
			bm.liveBullets = bm.liveBullets[:len(bm.liveBullets)-1]

			// Return bullet to pool
			bm.pool.PutBullet(bulletToRemove)
			break
		}
	}
}

// GetLivingBullets returns all living bullets (creates a copy)
func (bm *BulletManager) GetLivingBullets() []*Bullet {
	// Create a copy to avoid external modification
	result := make([]*Bullet, len(bm.liveBullets))
	copy(result, bm.liveBullets)
	return result
}

// GetBulletsNear returns bullets near the given position
func (bm *BulletManager) GetBulletsNear(x, y, radius int) []*Bullet {
	return bm.spatialIndex.QueryArea(x, y, radius)
}

// TrimMemory reduces memory usage by trimming slice capacities
func (bm *BulletManager) TrimMemory() {
	// If live bullets slice has excessive capacity, create a new one
	if cap(bm.liveBullets) > len(bm.liveBullets)*2 && cap(bm.liveBullets) > 100 {
		newSlice := make([]*Bullet, len(bm.liveBullets))
		copy(newSlice, bm.liveBullets)
		bm.liveBullets = newSlice
	}

	// Also clean up the object pool if it's too large
	if len(bm.pool.bullets) > bm.pool.maxSize/2 {
		bm.pool.bullets = bm.pool.bullets[:bm.pool.maxSize/2]
	}
}

// Clear removes all bullets
func (bm *BulletManager) Clear() {
	// Return all bullets to pool
	for _, bullet := range bm.liveBullets {
		bm.pool.PutBullet(bullet)
	}

	bm.liveBullets = bm.liveBullets[:0]
	bm.spatialIndex.Clear()
}

// GetStats returns performance statistics
func (bm *BulletManager) GetStats() (liveBullets, deadBullets int) {
	// Count bullets in pool as "available for reuse" rather than "dead"
	poolSize := 0
	poolSize = len(bm.pool.bullets)

	return len(bm.liveBullets), poolSize
}

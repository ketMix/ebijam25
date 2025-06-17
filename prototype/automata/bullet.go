package main

import (
	"fmt"
	"sync/atomic"
)

// Bullet represents an active cell with behavior defined by rules
type Bullet struct {
	// Position
	X, Y int

	// State
	alive         int32 // Use atomic for thread-safe access (1 = alive, 0 = dead)
	lifetime      int32 // Current age in simulation steps
	maxLifetime   int32 // Maximum lifetime before auto-death
	displayFrames int32 // How many frames this bullet has been displayed

	// Behavior
	ruleSet *RuleSet // Rules that govern this bullet's behavior

	// Metadata
	id         int64 // Unique identifier
	generation int   // Generation number (for tracking ancestry)
	chainID    int64 // Chain ID for isolating CA calculations

	// Visual properties (for rendering)
	colorR, colorG, colorB uint8
}

// bulletIDCounter for generating unique IDs
var bulletIDCounter int64

// chainIDCounter for generating unique chain IDs
var chainIDCounter int64

// NewBullet creates a new bullet with the given position and rules (clones the ruleset)
func NewBullet(x, y int, ruleSet *RuleSet) *Bullet {
	// Always clone the rule set so bullets maintain their behavior independently
	var clonedRuleSet *RuleSet
	if ruleSet != nil {
		clonedRuleSet = ruleSet.Clone()
	}

	bullet := &Bullet{
		X:          x,
		Y:          y,
		ruleSet:    clonedRuleSet,
		generation: 0,
		chainID:    atomic.AddInt64(&chainIDCounter, 1), // Each bullet starts a new chain by default
		id:         atomic.AddInt64(&bulletIDCounter, 1),
		colorR:     255,
		colorG:     255,
		colorB:     255,
	}

	// Set default lifetime
	if clonedRuleSet != nil {
		bullet.maxLifetime = int32(clonedRuleSet.GetDefaultLifetime())
	} else {
		bullet.maxLifetime = 100
	}

	atomic.StoreInt32(&bullet.alive, 1)
	return bullet
}

// NewBulletWithChainID creates a new bullet with a specific chain ID
func NewBulletWithChainID(x, y int, ruleSet *RuleSet, chainID int64) *Bullet {
	// Always clone the rule set so bullets maintain their behavior independently
	var clonedRuleSet *RuleSet
	if ruleSet != nil {
		clonedRuleSet = ruleSet.Clone()
	}

	bullet := &Bullet{
		X:          x,
		Y:          y,
		ruleSet:    clonedRuleSet,
		generation: 0,
		chainID:    chainID,
		id:         atomic.AddInt64(&bulletIDCounter, 1),
		colorR:     255,
		colorG:     255,
		colorB:     255,
	}

	// Set default lifetime
	if clonedRuleSet != nil {
		bullet.maxLifetime = int32(clonedRuleSet.GetDefaultLifetime())
	} else {
		bullet.maxLifetime = 100
	}

	atomic.StoreInt32(&bullet.alive, 1)
	return bullet
}

// NewChainBullet creates a new bullet that starts a new chain
func NewChainBullet(x, y int, ruleSet *RuleSet) *Bullet {
	newChainID := atomic.AddInt64(&chainIDCounter, 1)
	return NewBulletWithChainID(x, y, ruleSet, newChainID)
}

// NewBulletWithLifetime creates a new bullet with a specific lifetime
func NewBulletWithLifetime(x, y int, ruleSet *RuleSet, maxLifetime int) *Bullet {
	bullet := NewBullet(x, y, ruleSet)
	bullet.maxLifetime = int32(maxLifetime)
	return bullet
}

// IsAlive returns true if the bullet is still alive (thread-safe)
func (b *Bullet) IsAlive() bool {
	return atomic.LoadInt32(&b.alive) == 1
}

// Kill marks the bullet as dead (thread-safe)
func (b *Bullet) Kill() {
	atomic.StoreInt32(&b.alive, 0)
}

// GetLifetime returns the current lifetime (thread-safe)
func (b *Bullet) GetLifetime() int {
	return int(atomic.LoadInt32(&b.lifetime))
}

// GetMaxLifetime returns the maximum lifetime (thread-safe)
func (b *Bullet) GetMaxLifetime() int {
	return int(atomic.LoadInt32(&b.maxLifetime))
}

// Age increases the bullet's lifetime by 1 and checks for death
func (b *Bullet) Age() {
	// Increment display frames to track how long bullet has been visible
	atomic.AddInt32(&b.displayFrames, 1)

	newLifetime := atomic.AddInt32(&b.lifetime, 1)
	maxLifetime := atomic.LoadInt32(&b.maxLifetime)

	// Kill bullet only after fade-out period is complete
	// Allow bullet to live beyond maxLifetime for fade-out effect
	fadeOutDuration := max(maxLifetime/4, 10)
	if newLifetime > maxLifetime+fadeOutDuration {
		b.Kill()
	}
}

// GetOpacity returns the bullet's opacity (0.0 to 1.0) based on its lifetime
func (b *Bullet) GetOpacity() float32 {
	lifetime := b.GetLifetime()
	maxLifetime := b.GetMaxLifetime()

	if maxLifetime <= 0 {
		return 1.0
	}

	// Full opacity during normal lifetime
	if lifetime <= maxLifetime {
		return 1.0
	}

	// Fade-out during extended lifetime
	fadeOutDuration := maxLifetime / 4
	if fadeOutDuration < 10 {
		fadeOutDuration = 10
	}

	fadeProgress := float32(lifetime-maxLifetime) / float32(fadeOutDuration)

	// Linear fade from 1.0 to 0.0 during fade-out period
	opacity := 1.0 - fadeProgress

	// Clamp between 0.0 and 1.0
	if opacity < 0.0 {
		return 0.0
	}
	if opacity > 1.0 {
		return 1.0
	}

	return opacity
}

// IsInFadeOut returns true if the bullet is in its fade-out phase
func (b *Bullet) IsInFadeOut() bool {
	lifetime := b.GetLifetime()
	maxLifetime := b.GetMaxLifetime()

	// Bullet is in fade-out if it has exceeded its functional lifetime
	// but is still alive (during the fade-out period)
	return lifetime > maxLifetime && b.IsAlive()
}

// GetLifetimeProgress returns how far through its lifetime the bullet is (0.0 to 1.0)
func (b *Bullet) GetLifetimeProgress() float32 {
	lifetime := b.GetLifetime()
	maxLifetime := b.GetMaxLifetime()

	if maxLifetime <= 0 {
		return 0.0
	}

	return float32(lifetime) / float32(maxLifetime)
}

// GetDisplayFrames returns how many frames this bullet has been displayed
func (b *Bullet) GetDisplayFrames() int {
	return int(atomic.LoadInt32(&b.displayFrames))
}

// SetMaxLifetime sets the maximum lifetime (thread-safe)
func (b *Bullet) SetMaxLifetime(maxLifetime int) {
	atomic.StoreInt32(&b.maxLifetime, int32(maxLifetime))
}

// GetRuleSet returns the bullet's rule set
func (b *Bullet) GetRuleSet() *RuleSet {
	return b.ruleSet
}

// SetRuleSet sets the bullet's rule set
func (b *Bullet) SetRuleSet(ruleSet *RuleSet) {
	b.ruleSet = ruleSet
}

// GetID returns the bullet's unique ID
func (b *Bullet) GetID() int64 {
	return b.id
}

// GetGeneration returns the bullet's generation
func (b *Bullet) GetGeneration() int {
	return b.generation
}

// SetGeneration sets the bullet's generation
func (b *Bullet) SetGeneration(generation int) {
	b.generation = generation
}

// GetColor returns the bullet's RGB color
func (b *Bullet) GetColor() (uint8, uint8, uint8) {
	return b.colorR, b.colorG, b.colorB
}

// SetColor sets the bullet's RGB color
func (b *Bullet) SetColor(r, g, blue uint8) {
	b.colorR = r
	b.colorG = g
	b.colorB = blue
}

// GetChainID returns the bullet's chain ID
func (b *Bullet) GetChainID() int64 {
	return b.chainID
}

// SetChainID sets the bullet's chain ID
func (b *Bullet) SetChainID(chainID int64) {
	b.chainID = chainID
}

// Clone creates a copy of the bullet with a new ID
func (b *Bullet) Clone() *Bullet {
	newBullet := &Bullet{
		X:          b.X,
		Y:          b.Y,
		ruleSet:    b.ruleSet,
		generation: b.generation + 1,
		chainID:    b.chainID, // Inherit chain ID from parent
		id:         atomic.AddInt64(&bulletIDCounter, 1),
		colorR:     b.colorR,
		colorG:     b.colorG,
		colorB:     b.colorB,
	}

	newBullet.maxLifetime = atomic.LoadInt32(&b.maxLifetime)
	atomic.StoreInt32(&newBullet.alive, 1)

	return newBullet
}

// String returns a string representation of the bullet
func (b *Bullet) String() string {
	lifetime := atomic.LoadInt32(&b.lifetime)
	maxLifetime := atomic.LoadInt32(&b.maxLifetime)
	alive := atomic.LoadInt32(&b.alive)

	return fmt.Sprintf("Bullet{ID:%d, Chain:%d, Pos:(%d,%d), Life:%d/%d, Alive:%t, Gen:%d}",
		b.id, b.chainID, b.X, b.Y, lifetime, maxLifetime, alive == 1, b.generation)
}

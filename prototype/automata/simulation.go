package main

import (
	"sync"
	"sync/atomic"
)

// SimulationEngine manages the cellular automata simulation
type SimulationEngine struct {
	grid          *Grid
	bulletManager *BulletManager
	ruleRegistry  *RuleRegistry

	// Performance settings
	maxActionsPerFrame int
	batchSize          int
	paused             bool
	stepCounter        int64
	frameCounter       int64

	// Visual settings
	ghostFadeTicks          int     // How many ticks ghosts fade over
	deathCooldownMultiplier float32 // Multiplier for death cooldown duration (0.0 = no cooldown, 1.0 = same as fade duration)

	// Action tracking for UI
	lastActionCount int
	lastQueuedCount int

	// Threading
	mutex sync.RWMutex

	// Action queue for batch processing
	actionQueue []Action
	queueMutex  sync.Mutex

	// Default rule set for new bullets
	defaultRuleSet *RuleSet
}

// NewSimulationEngine creates a new simulation engine
func NewSimulationEngine(width, height int) *SimulationEngine {
	return &SimulationEngine{
		grid:                    NewGrid(width, height),
		bulletManager:           NewBulletManager(width, height),
		ruleRegistry:            NewRuleRegistry(),
		maxActionsPerFrame:      2000, // Reduced from 5000 to 2000 for smoother performance
		batchSize:               25,   // Reduced from 50 to 25 for smaller processing chunks
		ghostFadeTicks:          30,   // Default 30 tick fade duration
		deathCooldownMultiplier: 1.0,  // Default 1.0 means cooldown = fade duration
		actionQueue:             make([]Action, 0, 200),
	}
}

// GetGrid returns the simulation grid
func (se *SimulationEngine) GetGrid() *Grid {
	return se.grid
}

// GetBulletManager returns the bullet manager
func (se *SimulationEngine) GetBulletManager() *BulletManager {
	return se.bulletManager
}

// GetRuleRegistry returns the rule registry
func (se *SimulationEngine) GetRuleRegistry() *RuleRegistry {
	return se.ruleRegistry
}

// AddBullet adds a new bullet to the simulation
func (se *SimulationEngine) AddBullet(x, y int, ruleSet *RuleSet) *Bullet {
	bullet := se.bulletManager.CreateBullet(x, y, ruleSet)
	se.grid.SetBulletAt(x, y, bullet)
	return bullet
}

// RemoveBullet removes a bullet from the simulation
func (se *SimulationEngine) RemoveBullet(x, y int) {
	bullet := se.grid.RemoveBulletAt(x, y)
	if bullet != nil {
		bullet.Kill()
		se.bulletManager.RemoveBullet(bullet)

		// Mark the cell as having a recent death for this chain
		cell := se.grid.GetCell(x, y)
		if cell != nil {
			// Calculate cooldown based on bullet's fade-out duration and multiplier
			maxLifetime := bullet.GetMaxLifetime()
			fadeOutDuration := maxLifetime / 4
			if fadeOutDuration < 10 {
				fadeOutDuration = 10
			}
			cooldownDuration := int(float32(fadeOutDuration) * se.deathCooldownMultiplier)
			cell.MarkDeath(se.stepCounter, cooldownDuration, bullet.GetChainID())
		}
	}
}

// SetPaused sets the pause state
func (se *SimulationEngine) SetPaused(paused bool) {
	se.paused = paused
}

// IsPaused returns the pause state
func (se *SimulationEngine) IsPaused() bool {
	return se.paused
}

// Update performs one simulation step
func (se *SimulationEngine) Update() {
	se.frameCounter++

	if se.paused {
		return
	}
	se.stepCounter++

	// Reset action tracking for this frame
	se.lastQueuedCount = 0

	// Update bullet manager (ages bullets, handles lifecycle)
	se.bulletManager.Update()

	// Sync grid with bullet manager BEFORE applying rules
	se.syncGridWithBullets()

	// Apply rules to all living bullets
	se.applyRulesToBullets()
	// Process action queue
	se.processActionQueue()
}

// applyRulesToBullets applies rules to all living bullets
func (se *SimulationEngine) applyRulesToBullets() {
	bullets := se.bulletManager.GetLivingBullets()
	var bulletsToRemove []*Bullet

	for _, bullet := range bullets {
		if !bullet.IsAlive() {
			bulletsToRemove = append(bulletsToRemove, bullet)
			continue
		}

		// Skip rule application for bullets in fade-out phase
		if bullet.IsInFadeOut() {
			continue
		}

		// Check if bullet has reached its functional step limit
		remainingSteps := bullet.GetMaxLifetime() - bullet.GetLifetime()
		if remainingSteps <= 0 {
			// Bullet has reached end of functional lifetime, but don't kill yet
			// Let it fade out naturally through the Age() method
			continue
		}
		ruleSet := bullet.GetRuleSet()
		// Skip bullets with no rules for now (let them age naturally)
		if ruleSet == nil {
			continue
		}

		// Skip rule matching check for now - let bullets age naturally
		// TODO: Re-enable rule-based death later
		// if !ruleSet.HasMatchingRule(se.grid, bullet.X, bullet.Y) {
		//	bullet.Kill()
		//	bulletsToRemove = append(bulletsToRemove, bullet)
		//	continue
		// }

		// Apply rules and collect actions
		actions := ruleSet.Apply(se.grid, bullet.X, bullet.Y)
		if actions != nil {
			se.queueActions(actions)
		}
	}

	// Remove killed bullets from bullet manager
	for _, bullet := range bulletsToRemove {
		se.bulletManager.RemoveBullet(bullet)
	}
}

// queueActions adds actions to the action queue
func (se *SimulationEngine) queueActions(actions []Action) {
	// If maxActionsPerFrame is -1, allow unlimited actions
	if se.maxActionsPerFrame == -1 {
		se.actionQueue = append(se.actionQueue, actions...)
		se.lastQueuedCount += len(actions)
		return
	}

	// Limit actions per frame to prevent performance issues
	remainingCapacity := se.maxActionsPerFrame - len(se.actionQueue)
	if remainingCapacity <= 0 {
		return
	}

	if len(actions) > remainingCapacity {
		actions = actions[:remainingCapacity]
	}

	se.actionQueue = append(se.actionQueue, actions...)

	// Track action counts for UI display
	se.lastQueuedCount += len(actions)
}

// processActionQueue processes all queued actions
func (se *SimulationEngine) processActionQueue() {
	actions := se.actionQueue
	se.actionQueue = se.actionQueue[:0] // Clear queue

	// Track how many actions we're processing this frame
	se.lastActionCount = len(actions)

	// Process actions in batches for better performance
	for i := 0; i < len(actions); i += se.batchSize {
		end := i + se.batchSize
		if end > len(actions) {
			end = len(actions)
		}

		se.processBatch(actions[i:end])
	}
}

// processBatch processes a batch of actions
func (se *SimulationEngine) processBatch(actions []Action) {
	for _, action := range actions {
		switch action.Type {
		case ActionCreateBullet:
			se.handleCreateBulletAction(action)
		case ActionKillBullet:
			se.handleKillBulletAction(action)
		case ActionMoveBullet:
			se.handleMoveBulletAction(action)
		case ActionModifyBullet:
			se.handleModifyBulletAction(action)
		}
	}
}

// handleCreateBulletAction handles bullet creation
func (se *SimulationEngine) handleCreateBulletAction(action Action) {
	if !se.grid.IsInBounds(action.X, action.Y) {
		return
	}

	// Determine rule set, lifetime, and chain ID for new bullet
	var ruleSet *RuleSet
	var maxLifetime int
	var chainID int64

	if action.Data != nil {
		// Handle new map-based data format
		if dataMap, ok := action.Data.(map[string]interface{}); ok {
			if rs, ok := dataMap["ruleSet"].(*RuleSet); ok {
				ruleSet = rs
			}
			if lifetime, ok := dataMap["maxLifetime"].(int); ok {
				maxLifetime = lifetime
			}
			if cid, ok := dataMap["chainID"].(int64); ok {
				chainID = cid
			}
		} else if rs, ok := action.Data.(*RuleSet); ok {
			// Handle old direct RuleSet format for compatibility
			ruleSet = rs
		}
	}

	// If no chain ID provided, create a new chain
	if chainID == 0 {
		chainID = atomic.AddInt64(&chainIDCounter, 1)
	}

	// Check if cell can accept a new bullet for this chain
	cell := se.grid.GetCell(action.X, action.Y)
	if cell == nil || !cell.CanAcceptNewBullet(se.stepCounter, chainID) {
		return // Cannot create bullet in this cell
	}
	// Create bullet with default rule set if none provided
	if ruleSet == nil {
		ruleSet = se.ruleRegistry.GetDefaultRuleSet()
	}

	// Create bullet properly using the bullet manager
	bullet := se.bulletManager.CreateBullet(action.X, action.Y, ruleSet)
	// Set the chain ID for the bullet
	bullet.SetChainID(chainID)
	se.grid.SetBulletAt(action.X, action.Y, bullet)

	if maxLifetime > 0 {
		bullet.SetMaxLifetime(maxLifetime)
	}
}

// handleKillBulletAction handles bullet killing
func (se *SimulationEngine) handleKillBulletAction(action Action) {
	cell := se.grid.GetCell(action.X, action.Y)
	if cell == nil || !cell.IsAlive() {
		return
	}

	// Allow killing - remove minimum display time restriction that was causing respawning
	se.RemoveBullet(action.X, action.Y)
}

// handleMoveBulletAction handles bullet movement
func (se *SimulationEngine) handleMoveBulletAction(action Action) {
	// Extract new position from action data
	if action.Data == nil {
		return
	}

	newPos, ok := action.Data.([2]int)
	if !ok {
		return
	}

	newX, newY := newPos[0], newPos[1]

	if !se.grid.IsInBounds(newX, newY) {
		return
	}

	// Get bullet from old position
	bullet := se.grid.RemoveBulletAt(action.X, action.Y)
	if bullet == nil {
		return
	}
	// Check if destination is free
	destCell := se.grid.GetCell(newX, newY)
	if destCell != nil && destCell.IsAlive() {
		// Destination occupied - bullet dies
		bullet.Kill()
		se.bulletManager.RemoveBullet(bullet)
		return
	}

	// Move bullet
	bullet.X = newX
	bullet.Y = newY
	se.grid.SetBulletAt(newX, newY, bullet)
}

// handleModifyBulletAction handles bullet modification
func (se *SimulationEngine) handleModifyBulletAction(action Action) {
	cell := se.grid.GetCell(action.X, action.Y)
	if cell == nil || !cell.IsAlive() {
		return
	}

	bullet := cell.Bullet

	// Apply modification based on action data
	if action.Data != nil {
		if modifier, ok := action.Data.(func(*Bullet)); ok {
			modifier(bullet)
		}
	}
}

// syncGridWithBullets ensures grid state matches bullet manager state
func (se *SimulationEngine) syncGridWithBullets() {
	// Clean up dead bullets from grid
	se.grid.ForEachCell(func(x, y int, cell *Cell) {
		if cell.Bullet != nil && !cell.Bullet.IsAlive() {
			// Mark cell as having a recent death before removing bullet
			maxLifetime := cell.Bullet.GetMaxLifetime()
			fadeOutDuration := maxLifetime / 4
			if fadeOutDuration < 10 {
				fadeOutDuration = 10
			}
			cooldownDuration := int(float32(fadeOutDuration) * se.deathCooldownMultiplier)
			cell.MarkDeath(se.stepCounter, cooldownDuration, cell.Bullet.GetChainID())

			// Remove dead bullet
			cell.RemoveBullet()
		}
	})
}

// Clear clears the entire simulation
func (se *SimulationEngine) Clear() {
	se.grid.Clear()
	se.bulletManager.Clear()

	se.actionQueue = se.actionQueue[:0]
}

// GetStats returns simulation statistics
func (se *SimulationEngine) GetStats() SimulationStats {
	liveBullets, deadBullets := se.bulletManager.GetStats()
	queuedActions := len(se.actionQueue)

	return SimulationStats{
		StepCounter:   se.stepCounter,
		FrameCounter:  se.frameCounter,
		LiveBullets:   liveBullets,
		DeadBullets:   deadBullets,
		QueuedActions: queuedActions,
		Paused:        se.paused,
	}
}

// GetLastActionCount returns the number of actions processed in the last frame
func (se *SimulationEngine) GetLastActionCount() int {
	return se.lastActionCount
}

// GetLastQueuedCount returns the number of actions queued in the last frame
func (se *SimulationEngine) GetLastQueuedCount() int {
	return se.lastQueuedCount
}

// GetGhostFadeTicks returns the current ghost fade duration
func (se *SimulationEngine) GetGhostFadeTicks() int {
	return se.ghostFadeTicks
}

// SetGhostFadeTicks sets the ghost fade duration
func (se *SimulationEngine) SetGhostFadeTicks(ticks int) {
	if ticks < 1 {
		ticks = 1 // Minimum 1 tick
	}
	se.ghostFadeTicks = ticks
}

// SetDeathCooldownMultiplier sets the death cooldown multiplier
func (se *SimulationEngine) SetDeathCooldownMultiplier(multiplier float32) {
	se.deathCooldownMultiplier = multiplier
}

// GetDeathCooldownMultiplier returns the death cooldown multiplier
func (se *SimulationEngine) GetDeathCooldownMultiplier() float32 {
	return se.deathCooldownMultiplier
}

// GetStepCounter returns the current step counter
func (se *SimulationEngine) GetStepCounter() int64 {
	return se.stepCounter
}

// SimulationStats contains simulation statistics
type SimulationStats struct {
	StepCounter   int64
	FrameCounter  int64
	LiveBullets   int
	DeadBullets   int
	QueuedActions int
	Paused        bool
}

// RuleRegistry manages rules and provides default rule sets
type RuleRegistry struct {
	rules          map[string]Rule
	ruleSets       map[string]*RuleSet
	defaultRuleSet *RuleSet
}

// NewRuleRegistry creates a new rule registry
func NewRuleRegistry() *RuleRegistry {
	registry := &RuleRegistry{
		rules:    make(map[string]Rule),
		ruleSets: make(map[string]*RuleSet),
	}

	// Create default rule set with basic Conway-like rules
	registry.defaultRuleSet = NewRuleSet("Default")

	return registry
}

// RegisterRule registers a rule with a unique name
func (rr *RuleRegistry) RegisterRule(name string, rule Rule) {
	rr.rules[name] = rule
}

// GetRule returns a rule by name
func (rr *RuleRegistry) GetRule(name string) Rule {
	return rr.rules[name]
}

// RegisterRuleSet registers a rule set with a unique name
func (rr *RuleRegistry) RegisterRuleSet(name string, ruleSet *RuleSet) {
	rr.ruleSets[name] = ruleSet
}

// GetRuleSet returns a rule set by name
func (rr *RuleRegistry) GetRuleSet(name string) *RuleSet {
	return rr.ruleSets[name]
}

// GetDefaultRuleSet returns the default rule set
func (rr *RuleRegistry) GetDefaultRuleSet() *RuleSet {
	return rr.defaultRuleSet.Clone()
}

// SetDefaultRuleSet sets the default rule set for new bullets
func (se *SimulationEngine) SetDefaultRuleSet(ruleSet *RuleSet) {
	se.defaultRuleSet = ruleSet
}

// GetDefaultRuleSet returns the default rule set for new bullets
func (se *SimulationEngine) GetDefaultRuleSet() *RuleSet {
	// Use the rule registry's default rule set
	return se.ruleRegistry.GetDefaultRuleSet()
}

// GetMaxActionsPerFrame returns the current max actions per frame limit
func (se *SimulationEngine) GetMaxActionsPerFrame() int {
	return se.maxActionsPerFrame
}

// SetMaxActionsPerFrame sets the max actions per frame limit
func (se *SimulationEngine) SetMaxActionsPerFrame(maxActions int) {
	if maxActions < -1 {
		maxActions = -1 // -1 means unlimited, minimum otherwise is 1
	} else if maxActions == 0 {
		maxActions = 1 // Don't allow 0, use 1 as minimum
	}
	se.maxActionsPerFrame = maxActions
}

// AddBulletWithDefaultRules adds a new bullet using the default rule set
func (se *SimulationEngine) AddBulletWithDefaultRules(x, y int) *Bullet {
	// Create a new chain for manually placed bullets
	newChainID := atomic.AddInt64(&chainIDCounter, 1)

	// Check if the cell can accept a new bullet for this new chain
	cell := se.grid.GetCell(x, y)
	if cell == nil || !cell.CanAcceptNewBullet(se.stepCounter, newChainID) {
		return nil // Cannot create bullet in this cell
	}

	ruleSet := se.GetDefaultRuleSet()
	if ruleSet == nil {
		// Create a basic "stay alive" rule set if no default is set
		ruleSet = se.createBasicRuleSet()
	}

	// Create bullet properly using the bullet manager
	bullet := se.bulletManager.CreateBullet(x, y, ruleSet)
	// Set the chain ID for manually placed bullets
	bullet.SetChainID(newChainID)
	se.grid.SetBulletAt(x, y, bullet)
	return bullet
}

// createBasicRuleSet creates a basic rule set for fallback
func (se *SimulationEngine) createBasicRuleSet() *RuleSet {
	ruleSet := NewRuleSet("Basic")
	rule := NewBasicRule("Stay Alive")

	// Pattern: alive cell stays alive
	condition := [3][3]CellState{
		{Any, Any, Any},
		{Any, Alive, Any},
		{Any, Any, Any},
	}
	result := [3][3]CellState{
		{Any, Any, Any},
		{Any, Alive, Any},
		{Any, Any, Any},
	}

	rule.AddPattern(NewPattern3x3(condition, result))
	ruleSet.AddRule(rule)
	return ruleSet
}

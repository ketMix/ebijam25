package main

import (
	"fmt"
	"slices"
)

// Rule interface defines the behavior of a cellular automata rule
type Rule interface {
	GetName() string
	GetPatterns() []Pattern3x3
	Apply(grid *Grid, x, y int) []Action
	MatchesAt(grid *Grid, x, y int) bool
}

// Action represents an action to be performed on the grid
type Action struct {
	Type ActionType
	X, Y int
	Data any // Additional data for the action
}

// ActionType defines the type of action
type ActionType int

const (
	ActionCreateBullet ActionType = iota
	ActionKillBullet
	ActionMoveBullet
	ActionModifyBullet
)

// BasicRule implements a simple pattern-matching rule
type BasicRule struct {
	name     string
	patterns []Pattern3x3
}

// NewBasicRule creates a new basic rule
func NewBasicRule(name string) *BasicRule {
	return &BasicRule{
		name:     name,
		patterns: make([]Pattern3x3, 0),
	}
}

// GetName returns the rule name
func (r *BasicRule) GetName() string {
	return r.name
}

// AddPattern adds a pattern to the rule
func (r *BasicRule) AddPattern(pattern Pattern3x3) {
	r.patterns = append(r.patterns, pattern)
}

// GetPatterns returns all patterns in the rule
func (r *BasicRule) GetPatterns() []Pattern3x3 {
	return slices.Clone(r.patterns) // Return a copy
}

// MatchesAt checks if any pattern matches at the given position
func (r *BasicRule) MatchesAt(grid *Grid, x, y int) bool {
	area := grid.GetArea3x3(x, y)
	for _, pattern := range r.patterns {
		if pattern.Matches(area) {
			return true
		}
	}
	return false
}

// Apply applies the rule at the given position and returns actions
func (r *BasicRule) Apply(grid *Grid, x, y int) []Action {
	area := grid.GetArea3x3(x, y)

	// Find the first matching pattern
	for _, pattern := range r.patterns {
		if pattern.Matches(area) {
			return r.generateActionsFromPattern(pattern, area, x, y, nil, nil)
		}
	}

	return nil
}

// ApplyWithRuleSet applies the rule and passes the rule set to new bullet actions
func (r *BasicRule) ApplyWithRuleSet(grid *Grid, x, y int, ruleSet *RuleSet) []Action {
	area := grid.GetArea3x3(x, y)

	// Find the first matching pattern
	for _, pattern := range r.patterns {
		if pattern.Matches(area) {
			return r.generateActionsFromPattern(pattern, area, x, y, ruleSet, nil)
		}
	}

	return nil
}

// ApplyWithParent applies the rule and passes parent bullet info for lifetime inheritance
func (r *BasicRule) ApplyWithParent(grid *Grid, x, y int, ruleSet *RuleSet, parentBullet *Bullet) []Action {
	area := grid.GetArea3x3(x, y)

	// Find the first matching pattern
	for _, pattern := range r.patterns {
		if pattern.Matches(area) {
			return r.generateActionsFromPattern(pattern, area, x, y, ruleSet, parentBullet)
		}
	}

	return nil
}

// generateActionsFromPattern creates actions based on pattern transformation
func (r *BasicRule) generateActionsFromPattern(pattern Pattern3x3, currentArea [3][3]CellState, centerX, centerY int, ruleSet *RuleSet, parentBullet *Bullet) []Action {
	var actions []Action
	result := pattern.Apply()

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := centerX+dx, centerY+dy

			currentState := currentArea[dy+1][dx+1]
			newState := result[dy+1][dx+1]

			// Skip Any states (no action needed)
			if newState == Any {
				continue
			} // Generate appropriate actions
			if currentState == Empty && newState == Alive {
				// Create action data that includes both rule set and parent lifetime
				actionData := map[string]interface{}{
					"ruleSet": ruleSet,
				}

				// If we have a parent bullet, inherit its chain ID and pass its lifetime - 1
				if parentBullet != nil {
					parentLifetime := parentBullet.GetMaxLifetime() - parentBullet.GetLifetime()
					childLifetime := max(parentLifetime-1, 1)
					actionData["maxLifetime"] = childLifetime
					actionData["chainID"] = parentBullet.GetChainID() // Children inherit parent's chain ID
				}

				actions = append(actions, Action{
					Type: ActionCreateBullet,
					X:    nx,
					Y:    ny,
					Data: actionData,
				})
			} else if currentState == Alive && newState == Empty {
				// Kill existing bullet
				actions = append(actions, Action{
					Type: ActionKillBullet,
					X:    nx,
					Y:    ny,
				})
			}
		}
	}

	return actions
}

// RuleSet represents a collection of rules that define behavior
type RuleSet struct {
	name            string
	rules           []Rule
	defaultLifetime int
}

// NewRuleSet creates a new rule set
func NewRuleSet(name string) *RuleSet {
	return &RuleSet{
		name:            name,
		rules:           make([]Rule, 0),
		defaultLifetime: 100,
	}
}

// GetName returns the rule set name
func (rs *RuleSet) GetName() string {
	return rs.name
}

// AddRule adds a rule to the set
func (rs *RuleSet) AddRule(rule Rule) {
	rs.rules = append(rs.rules, rule)
}

// RemoveRule removes a rule from the set
func (rs *RuleSet) RemoveRule(rule Rule) {
	for i, r := range rs.rules {
		if r == rule {
			rs.rules = append(rs.rules[:i], rs.rules[i+1:]...)
			break
		}
	}
}

// GetRules returns all rules in the set
func (rs *RuleSet) GetRules() []Rule {
	return append([]Rule(nil), rs.rules...) // Return a copy
}

// GetDefaultLifetime returns the default lifetime for bullets with this rule set
func (rs *RuleSet) GetDefaultLifetime() int {
	return rs.defaultLifetime
}

// SetDefaultLifetime sets the default lifetime
func (rs *RuleSet) SetDefaultLifetime(lifetime int) {
	rs.defaultLifetime = lifetime
}

// Apply applies all rules at the given position and returns combined actions
func (rs *RuleSet) Apply(grid *Grid, x, y int) []Action {
	var allActions []Action

	// Get the bullet at this position to pass lifetime info to children
	cell := grid.GetCell(x, y)
	var parentBullet *Bullet
	if cell != nil && cell.IsAlive() {
		parentBullet = cell.Bullet
	}

	for _, rule := range rs.rules {
		// Use the new method that passes the parent bullet for lifetime inheritance
		if basicRule, ok := rule.(*BasicRule); ok {
			actions := basicRule.ApplyWithParent(grid, x, y, rs, parentBullet)
			if actions != nil {
				allActions = append(allActions, actions...)
			}
		} else {
			// Fallback for other rule types
			actions := rule.Apply(grid, x, y)
			if actions != nil {
				allActions = append(allActions, actions...)
			}
		}
	}

	return allActions
}

// HasMatchingRule returns true if any rule matches at the given position
func (rs *RuleSet) HasMatchingRule(grid *Grid, x, y int) bool {
	for _, rule := range rs.rules {
		if rule.MatchesAt(grid, x, y) {
			return true
		}
	}
	return false
}

// Clone creates a copy of the rule set
func (rs *RuleSet) Clone() *RuleSet {
	newRuleSet := NewRuleSet(rs.name)
	newRuleSet.defaultLifetime = rs.defaultLifetime

	// Share rules (they should be immutable)
	for _, rule := range rs.rules {
		newRuleSet.rules = append(newRuleSet.rules, rule)
	}

	return newRuleSet
}

// String returns a string representation of the rule set
func (rs *RuleSet) String() string {
	return fmt.Sprintf("RuleSet{name: %s, rules: %d, lifetime: %d}",
		rs.name, len(rs.rules), rs.defaultLifetime)
}

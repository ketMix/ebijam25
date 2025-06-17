package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	EditorX      = 600
	EditorY      = 10
	EditorWidth  = 280
	EditorHeight = 500
	GridSize     = 3
	CellSize     = 20
	GridSpacing  = 100 // Space between precedent and consequent grids
)

// RuleEditor provides a GUI for editing cellular automata rules with precedent/consequent
type RuleEditor struct {
	game            *Game
	currentRule     *BasicRule
	currentRuleSet  *RuleSet
	selectedPattern int

	// UI positions
	precedentGridX  int
	precedentGridY  int
	consequentGridX int
	consequentGridY int
}

// NewRuleEditor creates a new rule editor with precedent/consequent grids
func NewRuleEditor(game *Game) *RuleEditor {
	editor := &RuleEditor{
		game:            game,
		currentRule:     NewBasicRule("Custom Rule"),
		selectedPattern: 0,
		precedentGridX:  EditorX + 10,
		precedentGridY:  EditorY + 120,
		consequentGridX: EditorX + GridSpacing + 10,
		consequentGridY: EditorY + 120,
	}

	editor.currentRuleSet = NewRuleSet("Custom")
	editor.initializeDefaultPattern()

	return editor
}

func (re *RuleEditor) initializeDefaultPattern() {
	condition := [3][3]CellState{
		{Any, Any, Any},
		{Any, Alive, Any}, // Alive cell in center, any surrounding
		{Any, Any, Any},
	}
	result := [3][3]CellState{
		{Alive, Any, Alive},
		{Any, Empty, Any}, // Cell dies, creates bullets in corners
		{Alive, Any, Alive},
	}

	pattern := NewPattern3x3(condition, result)
	re.currentRule.AddPattern(pattern)
	re.currentRuleSet.AddRule(re.currentRule)

	// Set a shorter default lifetime for better visibility of the effect
	re.currentRuleSet.SetDefaultLifetime(30) // 30 simulation steps = about 1 second at normal speed
}

func (re *RuleEditor) Update() {
	re.handleInput()
}

// handleInput processes input for the rule editor
func (re *RuleEditor) handleInput() {
	// Add new pattern
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		re.addNewPattern()
	}

	// Remove current pattern
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		re.removeCurrentPattern()
	}

	// Navigate between patterns
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		if re.selectedPattern > 0 {
			re.selectedPattern--
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		patterns := re.currentRule.GetPatterns()
		if re.selectedPattern < len(patterns)-1 {
			re.selectedPattern++
		}
	}

	// Set cell states with number keys
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		re.setCellStateAtCursor(Empty)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		re.setCellStateAtCursor(Alive)
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		re.setCellStateAtCursor(Any)
	}

	// Create test bullet with current rules at center
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		re.createTestBullet()
	}

	// Clear all bullets
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		re.clearAllBullets()
	}

	// Performance controls - adjust max actions per frame
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		// Increase max actions per frame
		current := re.game.simulation.GetMaxActionsPerFrame()
		re.game.simulation.SetMaxActionsPerFrame(current + 100)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		// Decrease max actions per frame
		current := re.game.simulation.GetMaxActionsPerFrame()
		newValue := current - 100
		if newValue < 100 {
			newValue = 100
		}
		re.game.simulation.SetMaxActionsPerFrame(newValue)
	}

	// Toggle unlimited actions with backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		current := re.game.simulation.GetMaxActionsPerFrame()
		if current == -1 {
			// Currently unlimited, set back to a reasonable limit
			re.game.simulation.SetMaxActionsPerFrame(1000)
		} else {
			// Set to unlimited
			re.game.simulation.SetMaxActionsPerFrame(-1)
		}
	}

	// Bullet step limit controls
	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		// Decrease bullet step limit
		current := re.currentRuleSet.GetDefaultLifetime()
		newValue := current - 5
		if newValue < 5 {
			newValue = 5
		}
		re.currentRuleSet.SetDefaultLifetime(newValue)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		// Increase bullet step limit
		current := re.currentRuleSet.GetDefaultLifetime()
		newValue := current + 5
		if newValue > 500 {
			newValue = 500
		}
		re.currentRuleSet.SetDefaultLifetime(newValue)
	}

	// Handle mouse clicks on the grids
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		re.handleGridClick(x, y)
	}

	// Auto-apply rules whenever they change (no manual Enter needed)
	re.autoApplyRules()
}

// autoApplyRules automatically applies the current rule set to new bullets
func (re *RuleEditor) autoApplyRules() {
	// Always keep the default rule set updated for new bullets
	registry := re.game.simulation.GetRuleRegistry()
	registry.RegisterRuleSet("custom", re.currentRuleSet.Clone())
	re.game.simulation.SetDefaultRuleSet(re.currentRuleSet.Clone())
}

// setCellStateAtCursor sets the cell state at the cursor position on the detected grid
func (re *RuleEditor) setCellStateAtCursor(state CellState) {
	x, y := ebiten.CursorPosition()
	gridX, gridY, isConsequent, valid := re.screenToGrid(x, y)
	if valid {
		re.setCellState(gridX, gridY, state, isConsequent)
	}
}

// setCellState sets the state of a cell in the current pattern based on grid clicked
func (re *RuleEditor) setCellState(gridX, gridY int, state CellState, isConsequent bool) {
	patterns := re.currentRule.GetPatterns()
	if re.selectedPattern >= len(patterns) {
		return
	}

	pattern := patterns[re.selectedPattern]

	// Modify the appropriate grid based on which was clicked
	if isConsequent {
		pattern.Result[gridY][gridX] = state
	} else {
		pattern.Condition[gridY][gridX] = state
	}

	// Recreate the rule with all patterns
	newRule := NewBasicRule(re.currentRule.GetName())
	for i, p := range patterns {
		if i == re.selectedPattern {
			newRule.AddPattern(pattern)
		} else {
			newRule.AddPattern(p)
		}
	}

	re.currentRule = newRule
	re.updateRuleSet()
}

// updateRuleSet updates the current rule set with the current rule
func (re *RuleEditor) updateRuleSet() {
	re.currentRuleSet = NewRuleSet("Custom")
	re.currentRuleSet.AddRule(re.currentRule)
}

// addNewPattern adds a new empty pattern to the current rule
func (re *RuleEditor) addNewPattern() {
	condition := [3][3]CellState{
		{Empty, Empty, Empty},
		{Empty, Empty, Empty},
		{Empty, Empty, Empty},
	}
	result := [3][3]CellState{
		{Empty, Empty, Empty},
		{Empty, Empty, Empty},
		{Empty, Empty, Empty},
	}

	pattern := NewPattern3x3(condition, result)
	re.currentRule.AddPattern(pattern)
	re.selectedPattern = len(re.currentRule.GetPatterns()) - 1
	re.updateRuleSet()
}

// removeCurrentPattern removes the currently selected pattern
func (re *RuleEditor) removeCurrentPattern() {
	patterns := re.currentRule.GetPatterns()
	if len(patterns) <= 1 {
		return // Keep at least one pattern
	}

	newRule := NewBasicRule(re.currentRule.GetName())
	for i, pattern := range patterns {
		if i != re.selectedPattern {
			newRule.AddPattern(pattern)
		}
	}

	re.currentRule = newRule
	if re.selectedPattern >= len(re.currentRule.GetPatterns()) {
		re.selectedPattern = len(re.currentRule.GetPatterns()) - 1
	}
	re.updateRuleSet()
}

// BulletCreateAction represents a bullet creation action
type BulletCreateAction struct {
	X, Y        int
	RuleSetName string
}

// BulletKillAction represents a bullet kill action
type BulletKillAction struct {
	X, Y int
}

// createTestBullet creates a test bullet at the center of the grid with current rules
func (re *RuleEditor) createTestBullet() {
	centerX := GridWidth / 2
	centerY := GridHeight / 2

	// Apply current rule set to registry and set as default
	registry := re.game.simulation.GetRuleRegistry()
	registry.RegisterRuleSet("custom", re.currentRuleSet.Clone())
	re.game.simulation.SetDefaultRuleSet(re.currentRuleSet.Clone())

	// Create bullet with the custom rule set
	re.game.simulation.AddBullet(centerX, centerY, re.currentRuleSet)
}

// clearAllBullets removes all bullets from the simulation
func (re *RuleEditor) clearAllBullets() {
	bulletManager := re.game.simulation.GetBulletManager()
	bullets := bulletManager.GetLivingBullets()

	// Kill all bullets
	for _, bullet := range bullets {
		bullet.Kill()
	}

	// Clear the grid
	grid := re.game.simulation.GetGrid()
	grid.Clear()
}

// handleGridClick handles mouse clicks on the editing grids
func (re *RuleEditor) handleGridClick(screenX, screenY int) {
	gridX, gridY, isConsequent, valid := re.screenToGrid(screenX, screenY)
	if !valid {
		return
	}

	// Cycle through cell states on click
	patterns := re.currentRule.GetPatterns()
	if re.selectedPattern >= len(patterns) {
		return
	}

	pattern := patterns[re.selectedPattern]
	var currentState CellState

	if isConsequent {
		currentState = pattern.Result[gridY][gridX]
	} else {
		currentState = pattern.Condition[gridY][gridX]
	}

	// Cycle: Empty -> Alive -> Any -> Empty
	var newState CellState
	switch currentState {
	case Empty:
		newState = Alive
	case Alive:
		newState = Any
	case Any:
		newState = Empty
	}

	re.setCellState(gridX, gridY, newState, isConsequent)
}

// isInPrecedentGrid checks if screen coordinates are in the precedent grid
func (re *RuleEditor) isInPrecedentGrid(screenX, screenY int) bool {
	return screenX >= re.precedentGridX &&
		screenX < re.precedentGridX+GridSize*CellSize &&
		screenY >= re.precedentGridY &&
		screenY < re.precedentGridY+GridSize*CellSize
}

// isInConsequentGrid checks if screen coordinates are in the consequent grid
func (re *RuleEditor) isInConsequentGrid(screenX, screenY int) bool {
	return screenX >= re.consequentGridX &&
		screenX < re.consequentGridX+GridSize*CellSize &&
		screenY >= re.consequentGridY &&
		screenY < re.consequentGridY+GridSize*CellSize
}

// screenToGrid converts screen coordinates to grid coordinates and determines which grid
func (re *RuleEditor) screenToGrid(screenX, screenY int) (gridX, gridY int, isConsequent, valid bool) {
	var localX, localY int
	isConsequent = false

	if re.isInPrecedentGrid(screenX, screenY) {
		localX = screenX - re.precedentGridX
		localY = screenY - re.precedentGridY
		isConsequent = false
	} else if re.isInConsequentGrid(screenX, screenY) {
		localX = screenX - re.consequentGridX
		localY = screenY - re.consequentGridY
		isConsequent = true
	} else {
		return 0, 0, false, false
	}

	if localX < 0 || localY < 0 {
		return 0, 0, false, false
	}

	gridX = localX / CellSize
	gridY = localY / CellSize

	valid = gridX >= 0 && gridY >= 0 && gridX < GridSize && gridY < GridSize
	return gridX, gridY, isConsequent, valid
}

// Draw draws the rule editor with precedent and consequent grids
func (re *RuleEditor) Draw(screen *ebiten.Image) {
	// Draw background panel
	vector.DrawFilledRect(screen, EditorX, EditorY, EditorWidth, EditorHeight,
		color.RGBA{40, 40, 50, 255}, false)

	// Draw border
	vector.StrokeRect(screen, EditorX, EditorY, EditorWidth, EditorHeight, 2,
		color.RGBA{100, 100, 120, 255}, false)

	// Draw title
	text.Draw(screen, "CA Rule Editor", basicfont.Face7x13, EditorX+10, EditorY+20,
		color.RGBA{255, 255, 255, 255})

	// Draw grid labels
	text.Draw(screen, "Precedent", basicfont.Face7x13, re.precedentGridX, re.precedentGridY-10,
		color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "Consequent", basicfont.Face7x13, re.consequentGridX, re.consequentGridY-10,
		color.RGBA{200, 200, 200, 255})

	// Draw arrow between grids
	arrowY := re.precedentGridY + GridSize*CellSize/2
	arrowStartX := re.precedentGridX + GridSize*CellSize + 10
	arrowEndX := re.consequentGridX - 10
	re.drawArrow(screen, arrowStartX, arrowY, arrowEndX, arrowY)

	// Draw pattern info
	patterns := re.currentRule.GetPatterns()
	patternText := fmt.Sprintf("Pattern %d/%d", re.selectedPattern+1, len(patterns))
	text.Draw(screen, patternText, basicfont.Face7x13, EditorX+10, EditorY+40,
		color.RGBA{200, 200, 200, 255})

	// Draw the grids
	re.drawPrecedentGrid(screen)
	re.drawConsequentGrid(screen)

	// Draw instructions
	re.drawInstructions(screen)
}

// drawArrow draws an arrow from start to end coordinates
func (re *RuleEditor) drawArrow(screen *ebiten.Image, startX, startY, endX, endY int) {
	// Draw line
	vector.StrokeLine(screen, float32(startX), float32(startY), float32(endX), float32(endY), 2,
		color.RGBA{200, 200, 200, 255}, false)

	// Draw arrowhead
	arrowSize := 5
	vector.DrawFilledRect(screen, float32(endX-arrowSize), float32(endY-arrowSize/2),
		float32(arrowSize), float32(arrowSize), color.RGBA{200, 200, 200, 255}, false)
}

// drawPrecedentGrid draws the precedent (condition) grid
func (re *RuleEditor) drawPrecedentGrid(screen *ebiten.Image) {
	patterns := re.currentRule.GetPatterns()
	if re.selectedPattern >= len(patterns) {
		return
	}
	pattern := patterns[re.selectedPattern]
	grid := pattern.Condition

	re.drawGrid(screen, grid, re.precedentGridX, re.precedentGridY, false)
}

// drawConsequentGrid draws the consequent (result) grid
func (re *RuleEditor) drawConsequentGrid(screen *ebiten.Image) {
	patterns := re.currentRule.GetPatterns()
	if re.selectedPattern >= len(patterns) {
		return
	}

	pattern := patterns[re.selectedPattern]
	grid := pattern.Result

	re.drawGrid(screen, grid, re.consequentGridX, re.consequentGridY, false)
}

func (re *RuleEditor) drawGrid(screen *ebiten.Image, grid [3][3]CellState, offsetX, offsetY int, isSelected bool) {
	for y := 0; y < GridSize; y++ {
		for x := 0; x < GridSize; x++ {
			cellX := float32(offsetX + x*CellSize)
			cellY := float32(offsetY + y*CellSize)

			// Choose color based on cell state
			var cellColor color.RGBA
			switch grid[y][x] {
			case Empty:
				cellColor = color.RGBA{60, 60, 70, 255}
			case Alive:
				cellColor = color.RGBA{100, 255, 100, 255}
			case Any:
				cellColor = color.RGBA{255, 255, 100, 255}
			}

			// Draw cell
			vector.DrawFilledRect(screen, cellX, cellY, CellSize-2, CellSize-2, cellColor, false)

			// Draw border (highlight if selected grid)
			borderColor := color.RGBA{150, 150, 150, 255}
			if isSelected {
				borderColor = color.RGBA{255, 255, 255, 255}
			}
			vector.StrokeRect(screen, cellX, cellY, CellSize-2, CellSize-2, 1, borderColor, false)
		}
	}
}

func (re *RuleEditor) drawInstructions(screen *ebiten.Image) {
	y := EditorY + 250
	textColor := color.RGBA{180, 180, 180, 255}

	text.Draw(screen, "Instructions:", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 15
	text.Draw(screen, "Click grids: Edit cells", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "1: Empty  2: Alive  3: Any", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "←→: Navigate patterns", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "N: New pattern", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "Del: Remove pattern", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "Space: Create test bullet", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	text.Draw(screen, "C: Clear all", basicfont.Face7x13, EditorX+10, y, textColor)
	y += 24

	maxActions := re.game.simulation.GetMaxActionsPerFrame()
	processedActions := re.game.simulation.GetLastActionCount()
	queuedActions := re.game.simulation.GetLastQueuedCount()

	var perfText string
	if maxActions == -1 {
		perfText = "(+/-) Max Actions: UNLIMITED"
	} else {
		perfText = fmt.Sprintf("(+/-) Max Actions: %d", maxActions)
	}
	text.Draw(screen, perfText, basicfont.Face7x13, EditorX+10, y, textColor)
	y += 12
	actionText := fmt.Sprintf("Actions: %d p/ %d", processedActions, queuedActions)
	text.Draw(screen, actionText, basicfont.Face7x13, EditorX+10, y, textColor)
	y += 24

	// Bullet step limit controls
	lifetime := re.currentRuleSet.GetDefaultLifetime()
	lifetimeText := fmt.Sprintf("([/])Step Limit: %d", lifetime)
	text.Draw(screen, lifetimeText, basicfont.Face7x13, EditorX+10, y, textColor)
}

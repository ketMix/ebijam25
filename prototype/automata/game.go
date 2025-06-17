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

// Game represents the main game state
type Game struct {
	simulation *SimulationEngine
	ui         *UI
	input      *InputHandler
	renderer   *Renderer

	// Settings
	ticksPerSecond int
	pixelsPerCell  int
	showDebugInfo  bool

	// State
	frameCount int64
	lastMouseX int
	lastMouseY int
}

// NewGame creates a new game instance
func NewGame() *Game {
	simulation := NewSimulationEngine(GridWidth, GridHeight)

	game := &Game{
		simulation:     simulation,
		ticksPerSecond: 240, // 240 simulation steps per second
		pixelsPerCell:  6,
		showDebugInfo:  false,
	}

	game.ui = NewUI(game)
	game.input = NewInputHandler(game)
	game.renderer = NewRenderer(game)

	return game
}

// Update implements ebiten.Game
func (g *Game) Update() error {
	g.frameCount++

	// Handle input
	g.input.Update()

	g.simulation.Update()

	// Update UI
	g.ui.Update()

	return nil
}

// Draw implements ebiten.Game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen
	screen.Fill(color.RGBA{20, 20, 30, 255})

	// Render simulation
	g.renderer.DrawSimulation(screen)

	// Render UI
	g.ui.Draw(screen)
}

// Layout implements ebiten.Game
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}

// GetSimulation returns the simulation engine
func (g *Game) GetSimulation() *SimulationEngine {
	return g.simulation
}

// UI manages the user interface
type UI struct {
	game       *Game
	editor     *RuleEditor
	showEditor bool
	showStats  bool
}

// NewUI creates a new UI
func NewUI(game *Game) *UI {
	return &UI{
		game:       game,
		editor:     NewRuleEditor(game),
		showEditor: true,
		showStats:  true,
	}
}

// Update updates the UI
func (ui *UI) Update() {
	if ui.showEditor {
		ui.editor.Update()
	}
}

// Draw draws the UI
func (ui *UI) Draw(screen *ebiten.Image) {
	if ui.showStats {
		ui.drawStats(screen)
	}

	if ui.showEditor {
		ui.editor.Draw(screen)
	}

	ui.drawControls(screen)
}

// drawStats draws simulation statistics
func (ui *UI) drawStats(screen *ebiten.Image) {
	stats := ui.game.simulation.GetStats()

	y := 10
	textColor := color.RGBA{255, 255, 255, 255}

	text.Draw(screen, fmt.Sprintf("Steps: %d", stats.StepCounter), basicfont.Face7x13, 10, y, textColor)
	y += 15
	text.Draw(screen, fmt.Sprintf("Bullets: %d", stats.LiveBullets), basicfont.Face7x13, 10, y, textColor)
	y += 15
	text.Draw(screen, fmt.Sprintf("TPS: %d", ui.game.ticksPerSecond), basicfont.Face7x13, 10, y, textColor)
	y += 15
	text.Draw(screen, fmt.Sprintf("Death Cooldown: %.2f", ui.game.simulation.GetDeathCooldownMultiplier()), basicfont.Face7x13, 10, y, textColor)
	y += 15

	if stats.Paused {
		text.Draw(screen, "PAUSED", basicfont.Face7x13, 10, y, color.RGBA{255, 100, 100, 255})
	} else {
		text.Draw(screen, "RUNNING", basicfont.Face7x13, 10, y, color.RGBA{100, 255, 100, 255})
	}
}

// drawControls draws control instructions
func (ui *UI) drawControls(screen *ebiten.Image) {
	y := ScreenHeight - 190
	textColor := color.RGBA{200, 200, 200, 255}

	text.Draw(screen, "Controls:", basicfont.Face7x13, 10, y, textColor)
	y += 15
	if ui.showEditor {
		text.Draw(screen, "Editor Mode:", basicfont.Face7x13, 10, y, color.RGBA{100, 255, 100, 255})
		y += 12
		text.Draw(screen, "SPACE - Create test bullet", basicfont.Face7x13, 10, y, textColor)
		y += 12
		text.Draw(screen, "Click - Add bullet", basicfont.Face7x13, 10, y, textColor)
		y += 12
		text.Draw(screen, "Arrow keys - Move cursor", basicfont.Face7x13, 10, y, textColor)
		y += 12
		text.Draw(screen, "1/2/3 - Set cell state", basicfont.Face7x13, 10, y, textColor)
	} else {
		text.Draw(screen, "Game Mode:", basicfont.Face7x13, 10, y, color.RGBA{100, 255, 100, 255})
		y += 12
		text.Draw(screen, "P - Pause/Resume", basicfont.Face7x13, 10, y, textColor)
		y += 12
		text.Draw(screen, "Click - Add bullet", basicfont.Face7x13, 10, y, textColor)
		y += 12
		text.Draw(screen, "-/+ - Death cooldown", basicfont.Face7x13, 10, y, textColor)
	}

	y += 12
	text.Draw(screen, "C - Clear simulation", basicfont.Face7x13, 10, y, textColor)
	y += 12
	text.Draw(screen, "E - Toggle editor", basicfont.Face7x13, 10, y, textColor)
	y += 12
	text.Draw(screen, "D - Toggle debug info", basicfont.Face7x13, 10, y, textColor)
	y += 12
	text.Draw(screen, "[/] - Adjust TPS", basicfont.Face7x13, 10, y, textColor)
}

// InputHandler manages user input
type InputHandler struct {
	game *Game
}

// NewInputHandler creates a new input handler
func NewInputHandler(game *Game) *InputHandler {
	return &InputHandler{game: game}
}

// Update processes input
func (ih *InputHandler) Update() {
	// Toggle editor (always available)
	if inpututil.IsKeyJustPressed(ebiten.KeyE) {
		ih.game.ui.showEditor = !ih.game.ui.showEditor
	}

	// Toggle debug info (always available)
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		ih.game.showDebugInfo = !ih.game.showDebugInfo
	}

	// If editor is active, let it handle most inputs
	if ih.game.ui.showEditor {
		// Only allow these global controls when editor is active
		// Clear simulation
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			ih.game.simulation.Clear()
		}

		// Adjust TPS
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			if ih.game.ticksPerSecond > 30 {
				ih.game.ticksPerSecond -= 30
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			ih.game.ticksPerSecond += 30
		}
	} else {
		// Main game controls (only when editor is not active)
		// Pause/Resume
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			ih.game.simulation.SetPaused(!ih.game.simulation.IsPaused())
		}

		// Clear simulation
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			ih.game.simulation.Clear()
		}

		// Adjust TPS
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			if ih.game.ticksPerSecond > 30 {
				ih.game.ticksPerSecond -= 30
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			ih.game.ticksPerSecond += 30
		}

		// Adjust death cooldown multiplier
		if inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
			current := ih.game.simulation.GetDeathCooldownMultiplier()
			new := current - 0.25
			if new < 0.0 {
				new = 0.0
			}
			ih.game.simulation.SetDeathCooldownMultiplier(new)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
			current := ih.game.simulation.GetDeathCooldownMultiplier()
			new := current + 0.25
			if new > 3.0 {
				new = 3.0
			}
			ih.game.simulation.SetDeathCooldownMultiplier(new)
		}
	}
	// Mouse input for adding bullets (available in both modes)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		ih.handleMouseClick(x, y)
	}
}

// handleMouseClick handles mouse clicks
func (ih *InputHandler) handleMouseClick(screenX, screenY int) {
	// Convert screen coordinates to grid coordinates
	gridX := screenX / ih.game.pixelsPerCell
	gridY := screenY / ih.game.pixelsPerCell

	// Check if click is within simulation area
	if gridX >= 0 && gridY >= 0 && gridX < GridWidth && gridY < GridHeight {
		// If editor is active, use the current rule set from the editor
		if ih.game.ui.showEditor {
			// Get the current rule set from the editor
			currentRuleSet := ih.game.ui.editor.currentRuleSet

			// Apply current rule set to registry and set as default (same as spacebar)
			registry := ih.game.simulation.GetRuleRegistry()
			registry.RegisterRuleSet("custom", currentRuleSet.Clone())
			ih.game.simulation.SetDefaultRuleSet(currentRuleSet.Clone())

			// Create bullet with the custom rule set at clicked position
			ih.game.simulation.AddBullet(gridX, gridY, currentRuleSet)
		} else {
			// If not in editor mode, use default rules
			ih.game.simulation.AddBulletWithDefaultRules(gridX, gridY)
		}
	}
}

// Renderer handles drawing the simulation
type Renderer struct {
	game *Game
}

// NewRenderer creates a new renderer
func NewRenderer(game *Game) *Renderer {
	return &Renderer{game: game}
}

// DrawSimulation draws the simulation grid and bullets
func (r *Renderer) DrawSimulation(screen *ebiten.Image) {
	grid := r.game.simulation.GetGrid()
	pixelSize := float32(r.game.pixelsPerCell)
	currentStep := r.game.simulation.GetStepCounter()
	// Draw grid background
	gridColor := color.RGBA{40, 40, 50, 255}
	cooldownColor := color.RGBA{60, 30, 30, 255} // Slightly red tint for cooldown cells

	for x := 0; x < GridWidth; x++ {
		for y := 0; y < GridHeight; y++ {
			cell := grid.GetCell(x, y)
			cellColor := gridColor

			// Show death cooldown with a different background color
			if cell != nil && cell.HasAnyCooldown(currentStep) {
				cellColor = cooldownColor
			}

			vector.DrawFilledRect(screen,
				float32(x)*pixelSize, float32(y)*pixelSize,
				pixelSize-1, pixelSize-1,
				cellColor, false)
		}
	}

	// Draw bullets
	grid.ForEachCell(func(x, y int, cell *Cell) {
		// Only draw living bullets
		if cell.IsAlive() {
			bullet := cell.Bullet
			r, g, b := bullet.GetColor()

			// Simple test: make bullets more red as they age
			lifetime := bullet.GetLifetime()
			maxLifetime := bullet.GetMaxLifetime()
			if maxLifetime > 0 {
				ageRatio := float32(lifetime) / float32(maxLifetime)
				r = min(uint8(float32(r)+ageRatio*255), 255)
			}

			// Calculate alpha based on bullet lifetime (linear fade)
			opacity := bullet.GetOpacity()
			alpha := uint8(opacity * 255)

			bulletColor := color.RGBA{r, g, b, alpha}

			vector.DrawFilledRect(screen,
				float32(x)*pixelSize, float32(y)*pixelSize,
				pixelSize-1, pixelSize-1,
				bulletColor, false)
		}
	})

	// Draw debug info if enabled
	if r.game.showDebugInfo {
		r.drawDebugInfo(screen)
	}
}

// drawDebugInfo draws debug information
func (r *Renderer) drawDebugInfo(screen *ebiten.Image) {
	stats := r.game.simulation.GetStats()

	y := 200
	debugColor := color.RGBA{255, 255, 0, 255}

	text.Draw(screen, "=== DEBUG INFO ===", basicfont.Face7x13, 10, y, debugColor)
	y += 15
	text.Draw(screen, fmt.Sprintf("Frame: %d", stats.FrameCounter), basicfont.Face7x13, 10, y, debugColor)
	y += 12
	text.Draw(screen, fmt.Sprintf("Actions: %d", stats.QueuedActions), basicfont.Face7x13, 10, y, debugColor)
	y += 12
	text.Draw(screen, fmt.Sprintf("Dead: %d", stats.DeadBullets), basicfont.Face7x13, 10, y, debugColor)
}

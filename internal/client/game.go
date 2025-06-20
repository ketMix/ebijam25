package client

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/ketMix/ebijam25/internal/log"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Game represents the client-side game state and logic.
type Game struct {
	Joiner
	world.State
	log            *slog.Logger
	continentImage *ebiten.Image
	Debug          bool
	// NOTE: This will be removed if we switch to storing all schlub data in the ID.
	pendingConstituents pendingConstituentsList
	Constituents        []world.Constituent // oof.
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.log = log.New("game", "client")
	g.EventBus = *event.NewBus("client")

	// **** Event -> local state change hooks.
	g.EventBus.Subscribe((event.MetaJoin{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaJoin)
		fmt.Println("Player joined:", evt.Username, "ID:", evt.ID)
	})
	g.EventBus.Subscribe((event.MetaLeave{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaLeave)
		fmt.Println("Player left:", evt.ID)
	})
	g.EventBus.Subscribe((event.MetaWelcome{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaWelcome)
		g.PlayerID = evt.ID
		g.MobID = evt.MobID
		g.State.Continent = world.NewContinent(evt.Seed)
	})
	g.EventBus.Subscribe((event.MobSpawn{}).Type(), func(e event.Event) {
		evt := e.(*event.MobSpawn)
		if g.Continent == nil {
			g.log.Error("mob spawn event received but continent not initialized")
			return
		}

		mob := g.Continent.NewMob(evt.Owner, evt.ID, float64(evt.X), float64(evt.Y))

		// NOTE: This will be removed if we switch to storing all schlub data in the ID.
		// Check if we need constituents.
		for _, constituent := range evt.Constituents {
			if index := slices.IndexFunc(g.Constituents, func(c world.Constituent) bool {
				// for now...
				return c.(*world.Schlub).ID == constituent
			}); index == -1 {
				g.pendingConstituents.Add(evt.ID, constituent)
			} else {
				// Hey, we have it alreadie.
				mob.Constituents = append(mob.Constituents, g.Constituents[index])
			}
		}

		g.log.Debug("mob spawned", "id", evt.ID, "owner", evt.Owner, "x", evt.X, "y", evt.Y)
	})
	g.EventBus.Subscribe((event.MobDespawn{}).Type(), func(e event.Event) {
		evt := e.(*event.MobDespawn)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			g.Continent.RemoveMob(mob)
			g.log.Debug("mob despawned", "id", evt.ID)
		} else {
			g.log.Warn("mob despawned but not found", "id", evt.ID)
		}
	})
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			g.Continent.MoveMob(mob, float64(evt.X), float64(evt.Y))
			g.log.Debug("mob position updated", "id", evt.ID, "x", evt.X, "y", evt.Y)
		}
	})
	g.EventBus.Subscribe((event.MobMove{}).Type(), func(e event.Event) {
		evt := e.(*event.MobMove)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			mob.TargetX = float64(evt.X)
			mob.TargetY = float64(evt.Y)
			mob.TargetID = evt.TargetID
			g.log.Info("mob move requested", "id", evt.ID, "targetX", evt.X, "targetY", evt.Y, "targetID", evt.TargetID)
		}
	})
	// Schlubbin' NOTE: This will be removed if we switch to storing all schlub data in the ID.
	g.EventBus.Subscribe((event.SchlubCreateList{}).Type(), func(e event.Event) {
		evt := e.(*event.SchlubCreateList)
		for _, schlub := range evt.Schlubs {
			index := slices.IndexFunc(g.pendingConstituents, func(pc pendingMobConstituent) bool {
				return pc.Constituent == schlub.ID
			})
			if index != -1 {
				pending := g.pendingConstituents[index]
				// Remove pending.
				g.pendingConstituents = append(g.pendingConstituents[:index], g.pendingConstituents[index+1:]...)
				// Create new schlubbo.
				newSchlub := &world.Schlub{
					ID: schlub.ID,
				}
				if mob := g.Continent.Mobs.FindByID(pending.MobID); mob != nil {
					mob.Constituents = append(mob.Constituents, newSchlub)
					g.Constituents = append(g.Constituents, newSchlub)
					g.log.Debug("schlub created", "mobID", pending.MobID, "schlubID", schlub.ID)
				} else {
					g.log.Warn("schlub create event received but mob not found", "mobID", pending.MobID)
				}
			}
		}
	})

	// **** Request -> network send hooks.
	g.EventBus.Subscribe((request.Move{}).Type(), func(e event.Event) {
		// NOTE: We could do local interpolation here as well, so as to make the game feel more responsive in the event of lag.
		g.log.Debug("move request sent", "event", e)
	})
	g.EventBus.Subscribe((request.Leave{}).Type(), func(e event.Event) {
		g.log.Debug("leave request sent", "event", e)
	})
	g.EventBus.Subscribe((request.Construct{}).Type(), func(e event.Event) {
		g.log.Debug("construct request sent", "event", e)
	})
}

// Update updates the game state and processes events.
func (g *Game) Update() error {
	// Here is where we'd convert inputs, etc., into requests.
	// Just for testing.
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// For now, just send a move request to the server.
		x, y := ebiten.CursorPosition()
		g.EventBus.Publish(&request.Move{
			X: x,
			Y: y,
		})
		g.log.Debug("move request sent", "x", x, "y", y)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.Debug = !g.Debug
		g.log.Info("debug mode toggled", "enabled: ", g.Debug)
	}

	// Update the thingz.
	g.EventBus.ProcessEvents()

	/*for _, mob := range g.Mobs {
		mob.Update(&g.State)
	}*/
	return nil
}

func (g *Game) DrawDebug(screen *ebiten.Image) {
	if !g.Debug {
		return
	}

	y := func() func() int {
		currentY := 10
		return func() int {
			y := currentY
			currentY += 20
			return y
		}
	}()

	// Draw the comprehenisive debug info.
	// System Info
	x := 10
	ebitenutil.DebugPrintAt(screen, "System Info:\n", x-5, y())
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Screen Size: %dx%d", screen.Bounds().Dx(), screen.Bounds().Dy()), x, y())
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("FPS: %.2f | TPS: %.2f", ebiten.ActualFPS(), ebiten.ActualTPS()), x, y())
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Tickrate: %d", g.State.Tickrate), x, y())
	y()

	// Session Info
	ebitenutil.DebugPrintAt(screen, "Session Info:", x-5, y())
	if g.State.Continent == nil {
		ebitenutil.DebugPrintAt(screen, "Continent not initialized", x, y())
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Continent Seed: %d", g.State.Continent.Sneed), x, y())
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Continent Size: %d", len(g.Continent.Fiefs)), x, y())
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Player ID: %d | Mob ID: %d", g.PlayerID, g.MobID), x, y())
	}
	y()

	// Player Info
	p := g.Continent.Mobs.FindByID(g.MobID)
	ebitenutil.DebugPrintAt(screen, "Player Info:", x-5, y())
	if p == nil {
		ebitenutil.DebugPrintAt(screen, "Player not found", x, y())
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("X: %.2f | Y: %.2f", p.X, p.Y), x, y())
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Target X: %.2f | Target Y: %.2f", p.TargetX, p.TargetY), x, y())
	}
	mX, mY := ebiten.CursorPosition()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Cursor: (%d, %d)", mX, mY), x, y())
}

// Draw draws da game.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.continentImage == nil {
		return
	}

	g.continentImage.Clear()
	g.DrawContinent(g.continentImage)

	// Center image on player
	ops := &ebiten.DrawImageOptions{}
	if mob := g.Continent.Mobs.FindByID(g.MobID); mob != nil {
		ops.GeoM.Translate(-mob.X+float64(screen.Bounds().Dx()/2),
			-mob.Y+float64(screen.Bounds().Dy()/2))
	} else {
		g.log.Warn("draw called but mob not found", "mobID", g.MobID)
	}

	// Draw the image buffer to the screen.
	screen.DrawImage(g.continentImage, ops)
	if g.Debug {
		g.DrawDebug(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	// Return the original dimensions for now.
	if g.continentImage == nil || (g.continentImage.Bounds().Dx() != ow || g.continentImage.Bounds().Dy() != oh) {
		g.continentImage = ebiten.NewImage(ow, oh)
	}
	return ow, oh
}

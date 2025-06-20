package client

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
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
	log              *slog.Logger
	debug            Debug
	continentImage   *ebiten.Image
	cameraX, cameraY float64
	cameraLock       bool
	Debug            bool
	// NOTE: This will be removed if we switch to storing all schlub data in the ID.
	pendingConstituents pendingConstituentsList
	Constituents        []world.Constituent // oof.
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.log = log.New("game", "client")
	g.debug.Setup()
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

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.cameraLock = !g.cameraLock
		g.log.Info("camera lock toggled", "enabled: ", g.cameraLock)
	}

	// Move camera with WASD
	pressedKeys := inpututil.AppendPressedKeys(nil)
	y := 0.0
	x := 0.0
	mult := 1.0
	for _, key := range pressedKeys {
		switch key {
		case ebiten.KeyW, ebiten.KeyUp:
			y -= 1
		case ebiten.KeyS, ebiten.KeyDown:
			y += 1
		case ebiten.KeyA, ebiten.KeyLeft:
			x -= 1
		case ebiten.KeyD, ebiten.KeyRight:
			x += 1
		case ebiten.KeyShift:
			mult = 5.0 // Speed up camera movement with shift.
		}
	}

	if x != 0 || y != 0 {
		// Move the camera based on the pressed keys.
		g.cameraX += float64(x) * mult
		g.cameraY += float64(y) * mult
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

	systemString := "System Info:\n" +
		fmt.Sprintf(" Screen Size: %dx%d\n", screen.Bounds().Dx(), screen.Bounds().Dy()) +
		fmt.Sprintf(" FPS: %.2f | TPS: %.2f\n", ebiten.ActualFPS(), ebiten.ActualTPS()) +
		fmt.Sprintf(" Tickrate: %d\n", g.State.Tickrate) +
		"\n"

	sessionString := "Session Info:\n"
	if g.State.Continent == nil {
		sessionString += " Continent not initialized\n"
	} else {
		sessionString += fmt.Sprintf(" Continent Seed: %d\n", g.State.Continent.Sneed) +
			fmt.Sprintf(" Continent Size: %d\n", len(g.Continent.Fiefs)) +
			fmt.Sprintf(" Player ID: %d | Mob ID: %d\n", g.PlayerID, g.MobID) +
			"\n"
	}

	playerString := "Player Info:\n"
	if p := g.Continent.Mobs.FindByID(g.MobID); p == nil {
		playerString += " Player not found\n"
	} else {
		playerString += fmt.Sprintf(" X: %.2f | Y: %.2f\n", p.X, p.Y) +
			fmt.Sprintf(" Target X: %.2f | Target Y: %.2f\n", p.TargetX, p.TargetY) +
			"\n"
	}

	mX, mY := ebiten.CursorPosition()
	cursorString := fmt.Sprintf(" Cursor: (%d, %d)\n", mX, mY)

	g.debug.setLeftText(systemString + sessionString + playerString + cursorString)
}

// Draw draws da game.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.continentImage == nil {
		return
	}

	g.continentImage.Clear()
	g.DrawContinent(g.continentImage)

	ops := &ebiten.DrawImageOptions{}
	// Center image on player
	if g.cameraLock {
		mob := g.Continent.Mobs.FindByID(g.MobID)
		if mob != nil {
			g.cameraX = mob.X
			g.cameraY = mob.Y
		}
	}
	ops.GeoM.Translate(-g.cameraX+float64(screen.Bounds().Dx()/2),
		-g.cameraY+float64(screen.Bounds().Dy()/2))

	// Draw the image buffer to the screen.
	screen.DrawImage(g.continentImage, ops)
	if g.Debug {
		g.DrawDebug(screen)
		g.debug.Draw(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	// Return the original dimensions for now.
	if g.continentImage == nil || (g.continentImage.Bounds().Dx() != ow || g.continentImage.Bounds().Dy() != oh) {
		g.continentImage = ebiten.NewImage(ow, oh)
	}
	return ow, oh
}

package client

import (
	"fmt"
	"log/slog"

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
	log            *slog.Logger
	debug          Debug
	continentImage *ebiten.Image
	fiefImages     []*ebiten.Image
	cammie         Cammie
	Debug          bool
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.log = log.New("game", "client")
	g.debug.Setup()
	g.cammie.Setup()
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

		var schlubs []world.SchlubID
		for _, s := range evt.Schlubs {
			schlubs = append(schlubs, world.SchlubID(s))
		}

		mob := g.Continent.NewMob(evt.Owner, evt.ID, float64(evt.X), float64(evt.Y))
		mob.AddSchlub(schlubs...)

		g.log.Debug("mob spawned", "id", evt.ID, "owner", evt.Owner, "x", evt.X, "y", evt.Y, "schlubs", len(schlubs))
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
		// Convert screen coordinates to world coordinates.
		x, y := g.cammie.ScreenToWorld(ebiten.CursorPosition())
		g.EventBus.Publish(&request.Move{
			X: int(x),
			Y: int(y),
		})
		g.log.Debug("move request sent", "x", x, "y", y)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		g.Debug = !g.Debug
		g.log.Info("debug mode toggled", "enabled: ", g.Debug)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.cammie.ToggleLocked()
		g.log.Info("camera lock toggled", "enabled: ", g.cammie.Locked())
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
			mult = 5.0
		}
	}

	if x != 0 || y != 0 {
		// Move the camera based on the pressed keys.
		g.cammie.AddPosition(x*mult, y*mult)
	}
	g.EventBus.ProcessEvents()

	// Update the camera to reflect any positional changes.
	g.cammie.Update()

	// Update our debug info.
	if g.Debug {
		g.UpdateDebug()
	}

	/*for _, mob := range g.Mobs {
		mob.Update(&g.State)
	}*/
	return nil
}

func (g *Game) UpdateDebug() {
	systemString := "System Info:\n" +
		//fmt.Sprintf(" Screen Size: %dx%d\n", screen.Bounds().Dx(), screen.Bounds().Dy()) + // This doesn't seem necessary and due to me moving this to an update and not draw context, we don't have a screen here. Could re-add, ofc.
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
	if g.Continent == nil {
		playerString += " Continent not initialized\n"
	} else {
		if p := g.Continent.Mobs.FindByID(g.MobID); p == nil {
			playerString += " Player not found\n"
		} else {
			playerString += fmt.Sprintf(" X: %.2f | Y: %.2f\n", p.X, p.Y) +
				fmt.Sprintf(" Target X: %.2f | Target Y: %.2f\n", p.TargetX, p.TargetY) +
				"\n"
		}
	}

	mX, mY := ebiten.CursorPosition()
	worldX, worldY := g.cammie.ScreenToWorld(mX, mY)
	cursorString := fmt.Sprintf(" Cursor: (%d, %d)\n", mX, mY)
	cursorString += fmt.Sprintf(" World Coordinates: (%.2f, %.2f)\n", worldX, worldY)
	g.debug.setLeftText(systemString + sessionString + playerString + cursorString)
}

// Draw draws da game.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.continentImage == nil {
		return
	}

	// Draw ze continentie.
	g.continentImage.Clear()
	g.DrawContinent(g.continentImage)

	// Center camera on player
	if g.cammie.Locked() {
		mob := g.Continent.Mobs.FindByID(g.MobID)
		if mob != nil {
			g.cammie.SetPosition(mob.X, mob.Y)
		}
	}

	// Draw the continent to the camera.
	g.cammie.image.Clear()
	g.cammie.image.DrawImage(g.continentImage, &g.cammie.opts)

	// Draw the camera to the screen.
	g.cammie.Draw(screen)

	// And, of course, debuggies.
	if g.Debug {
		g.debug.Draw(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	// Return the original dimensions for now.
	if g.continentImage == nil || (g.continentImage.Bounds().Dx() != ow || g.continentImage.Bounds().Dy() != oh) {
		g.continentImage = ebiten.NewImage(ow, oh)
	}
	// Refresh the camera's image as necessary.
	g.cammie.Layout(ow, oh)
	return ow, oh
}

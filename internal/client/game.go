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
	players        []*world.Player
	log            *slog.Logger
	debug          Debug
	continentImage *ebiten.Image
	fiefImages     []*ebiten.Image
	cammie         Cammie
	Debug          bool
	Dialoggies     Dialoggies
	schlubSystem   map[world.ID]*Schlubs
	Joined         bool
	//
	skipTutorial       bool
	hasSeenFirstMob    bool
	hasSeenFirstPlayer bool
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.log = log.New("game", "client")
	g.debug.Setup()
	g.cammie.Setup()
	g.EventBus = *event.NewBus("client")
	g.continentImage = ebiten.NewImage(world.ContinentPixelSpan, world.ContinentPixelSpan)

	// **** Event -> local state change hooks.
	g.EventBus.Subscribe((event.MetaJoin{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaJoin)
		for _, player := range g.players {
			if player.ID == evt.ID {
				g.log.Warn("player already exists", "id", evt.ID, "username", evt.Username)
				return // Player already exists, no need to add again.
			}
		}
		g.players = append(g.players, world.NewPlayer(evt.Username, evt.ID, evt.Color))
	})
	g.EventBus.Subscribe((event.MetaLeave{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaLeave)
		fmt.Println("Player left:", evt.ID)
		for i, player := range g.players {
			if player.ID == evt.ID {
				g.players = append(g.players[:i], g.players[i+1:]...) // Remove the player from the slice.
				g.log.Info("player removed", "id", evt.ID, "username", player.Username)
				return
			}
		}
		g.log.Warn("player left but not found", "id", evt.ID)
	})
	g.EventBus.Subscribe((event.MetaWelcome{}).Type(), func(e event.Event) {
		evt := e.(*event.MetaWelcome)
		g.Color = evt.Color
		g.PlayerID = evt.ID

		// I guess we can presume a welcome event should proc adding the player.
		found := false
		for _, player := range g.players {
			if player.ID == evt.ID {
				found = true
				g.log.Warn("player already exists", "id", evt.ID, "username", evt.Username)
				break
			}
		}
		if !found {
			g.players = append(g.players, world.NewPlayer(evt.Username, evt.ID, evt.Color))
		}

		g.MobID = evt.MobID
		g.State.Continent = world.NewContinent(evt.Seed)
		g.State.Tickrate = evt.Rate
		g.Dialoggies.Add("SCHLUBWORLD", "Welcome to SCHLUBWORLD, "+evt.Username+"!\n\nIn this world, it is up to you to slowly rise to power by converting or defeating other schlubs!\nYour starting character must be kept alive.\n\nYour leader unit, henceforth known as \"you\" is very good at converting other schlubs, but be wary of other players or schlub mobs that are too large!", []string{"Skip Tutorials", "OK"}, func(s string) {
			if s == "Skip Tutorials" {
				g.skipTutorial = true
				g.log.Info("tutorial skipped")
			}
			g.Dialoggies.dialogs = g.Dialoggies.dialogs[1:] // Remove the dialog from the stack.
			g.Dialoggies.layout.ClearEvents()
			g.Dialoggies.Next()
		})
		g.Dialoggies.SetTitleColor(evt.Color) // Just for fanciness.
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
		g.schlubSystem[mob.ID] = NewSchlubs(mob)

		g.log.Debug("mob spawned", "id", evt.ID, "owner", evt.Owner, "x", evt.X, "y", evt.Y, "schlubs", len(schlubs))
		if mob.ID == g.MobID {
			player := g.Continent.Mobs.FindByID(g.MobID)
			if player != nil {
				g.cammie.SetPosition(player.X, player.Y)
			}
		} else {
			// If not this player's mob, show tutorial.
			if !g.skipTutorial {
				if !g.hasSeenFirstMob && mob.OwnerID == 0 && mob.ID != g.MobID {
					g.hasSeenFirstMob = true
					g.Dialoggies.Add("Mobs", "A new mob made up of random shlubs has appeared!\n\nThis may very well be the first mob you can convert by moving into it, but take care!\n\n", []string{"OK"}, func(s string) {
						g.Dialoggies.dialogs = g.Dialoggies.dialogs[1:] // Remove the dialog from the stack.
						g.Dialoggies.layout.ClearEvents()
						g.Dialoggies.Next()
					})
				}
				if !g.hasSeenFirstPlayer && mob.OwnerID != 0 && mob.ID != g.MobID {
					g.hasSeenFirstPlayer = true
					g.Dialoggies.Add("Players Mobs", "You've come in vision range of a new player!\n\nYou can see their mob on the map but they might not be able to see you if they're smaller.\n\nYou can convert or slay their schlubs by moving into them, but be careful! They may try to do the same to you!", []string{"OK"}, func(s string) {
						g.Dialoggies.dialogs = g.Dialoggies.dialogs[1:] // Remove the dialog from the stack.
						g.Dialoggies.layout.ClearEvents()
						g.Dialoggies.Next()
					})
				}
			}
		}
		// Finally, let's set the mob's color to the player's color.
		for _, player := range g.players {
			if player.ID == evt.Owner {
				mob.Color = player.Color
				g.log.Debug("mob color set", "id", evt.ID, "color", mob.Color)
				break
			}
		}
	})
	g.EventBus.Subscribe((event.MobDespawn{}).Type(), func(e event.Event) {
		evt := e.(*event.MobDespawn)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			g.Continent.RemoveMob(mob)
			// Remove particle system
			delete(g.schlubSystem, mob.ID)
			g.log.Debug("mob despawned", "id", evt.ID)
		} else {
			g.log.Warn("mob despawned but not found", "id", evt.ID)
		}
	})
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			floatX := float64(evt.X) / world.FloatScale
			floatY := float64(evt.Y) / world.FloatScale
			g.Continent.MoveMob(mob, floatX, floatY)
			g.log.Debug("mob position updated", "id", evt.ID, "x", floatX, "y", floatY)
		}
	})
	g.EventBus.Subscribe((event.MobMove{}).Type(), func(e event.Event) {
		evt := e.(*event.MobMove)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			floatX := float64(evt.X) / world.FloatScale
			floatY := float64(evt.Y) / world.FloatScale
			mob.TargetX = floatX
			mob.TargetY = floatY
			mob.TargetID = evt.TargetID
			g.log.Info("mob move requested", "id", evt.ID, "targetX", evt.X, "targetY", evt.Y, "targetID", evt.TargetID)
		}
	})
	g.EventBus.Subscribe((event.MobFormation{}).Type(), func(e event.Event) {
		evt := e.(*event.MobFormation)
		// FIXME: We should only check for mobs in the visual radius of the player.
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			// I guess find the matching schlubs since we have that as an extra abstraction now.
			if g.schlubSystem[mob.ID] != nil {
				g.schlubSystem[mob.ID].Swap(world.SchlubID(evt.OuterKind))
				g.log.Info("mob formation updated", "id", evt.ID, "formation", evt.OuterKind)
			} else {
				g.log.Warn("mob formation event received but schlub system not found for mob", "id", evt.ID)
			}
		}
	})
	g.EventBus.Subscribe((event.MobDamage{}).Type(), func(e event.Event) {
		evt := e.(*event.MobDamage)
		if mob := g.Continent.Mobs.FindByID(evt.ID); mob != nil {
			if len(evt.IDs) > 0 {
				var schlubs []world.SchlubID
				for _, id := range evt.IDs {
					schlubs = append(schlubs, world.SchlubID(id))
				}
				mob.RemoveSchlub(schlubs...)
				// Also remove from schlub system wart.
				if g.schlubSystem[mob.ID] != nil {
					g.schlubSystem[mob.ID].RemoveSchlubs(schlubs...)
				} else {
					g.log.Warn("mob damage event received but schlub system not found for mob", "id", evt.ID)
				}
				g.log.Info("mob damaged", "id", evt.ID, "schlubs removed", len(evt.IDs))
			} else {
				g.log.Warn("mob damage event received with no schlubs to remove", "id", evt.ID)
			}
		} else {
			g.log.Warn("mob damage event received but mob not found", "id", evt.ID)
		}
	})
	g.EventBus.Subscribe((event.MobConvert{}).Type(), func(e event.Event) {
		evt := e.(*event.MobConvert)
		var schlubs []world.SchlubID
		for _, id := range evt.IDs {
			schlubs = append(schlubs, world.SchlubID(id))
		}
		if fromMob := g.Continent.Mobs.FindByID(evt.From); fromMob != nil {
			if toMob := g.Continent.Mobs.FindByID(evt.To); toMob != nil {
				// Convert schlubs from one mob to another.
				if g.schlubSystem[fromMob.ID] != nil && g.schlubSystem[toMob.ID] != nil {
					collected := g.schlubSystem[fromMob.ID].CollectSchlubsByID(schlubs...)
					g.schlubSystem[toMob.ID].PersuadeSchlubs(collected)
					g.schlubSystem[fromMob.ID].RemoveSchlubs(schlubs...)
					g.log.Info("mob converted", "from", evt.From, "to", evt.To, "schlubs", len(evt.IDs))
				} else {
					g.log.Warn("mob convert event received but schlub system not found for one or both mobs", "from", evt.From, "to", evt.To)
				}
			} else {
				g.log.Warn("mob convert event received but to mob not found", "to", evt.To)
			}
		} else {
			g.log.Warn("mob convert event received but from mob not found", "from", evt.From)
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
	g.EventBus.Subscribe((request.Formation{}).Type(), func(e event.Event) {
		g.log.Debug("formation request sent", "event", e)
	})

	g.schlubSystem = make(map[world.ID]*Schlubs)
}

// Update updates the game state and processes events.
func (g *Game) Update() error {
	g.Dialoggies.Update()
	if !g.Joined {
		return nil
	}

	// Input handling (dialoggies do be blocking, though).
	if !g.Dialoggies.layout.HasEvents() && len(g.Dialoggies.dialogs) == 0 {
		// Here is where we'd convert inputs, etc., into requests.
		// Just for testing.
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			// Convert screen coordinates to world coordinates.
			x, y := g.cammie.ScreenToWorld(ebiten.CursorPosition())
			g.EventBus.Publish(&request.Move{
				X: int(x * world.FloatScale),
				Y: int(y * world.FloatScale),
			})
			g.log.Debug("move request sent", "x", x, "y", y)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			// Request a formation change for the player's mob.
			g.EventBus.Publish(&request.Formation{
				// Not populating for now...
			})
		}

		// Handle mouse wheel input for zooming.
		dX, dY := ebiten.Wheel()
		if dX != 0 || dY != 0 {
			if dY < 0 {
				g.cammie.UpdateZoom(-0.1) // Zoom out
			} else if dY > 0 {
				g.cammie.UpdateZoom(0.1) // Zoom in
			}
		}

		if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
			g.Debug = !g.Debug
			g.log.Info("debug mode toggled", "enabled: ", g.Debug)
		}

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.cammie.ToggleLocked()
			if g.cammie.Locked() {
				player := g.Continent.Mobs.FindByID(g.MobID)
				if player == nil {
					g.log.Error("camera lock failed: player not found", "mobID", g.MobID)
				} else {
					// If the camera is locked, center it on the player.
					g.cammie.SetPosition(player.X, player.Y)
				}
			}

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

	// Update schlubs
	for _, ps := range g.schlubSystem {
		ps.Update(float64(g.State.Tickrate))
	}
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
	g.DrawContinent(g.continentImage, g.cammie.zoom < 0.8)

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

	// Dialoggies.
	g.Dialoggies.Draw(screen)

	// And, of course, debuggies.
	if g.Debug {
		g.debug.Draw(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	g.Dialoggies.Layout(float64(ow), float64(oh))
	// Refresh the camera's image as necessary.
	g.cammie.Layout(ow, oh)
	return ow, oh
}

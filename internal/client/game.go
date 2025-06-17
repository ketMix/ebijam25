package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Game represents the client-side game state and logic.
type Game struct {
	world.State
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.EventBus = *event.NewBus()

	// **** Event -> local state change hooks.
	g.EventBus.Subscribe((event.MobSpawn{}).Type(), func(e event.Event) {
		evt := e.(*event.MobSpawn)
		mob := world.NewMob(evt.ID, evt.Owner, float64(evt.X), float64(evt.Y))
		g.Mobs.Add(mob)
	})
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.X = float64(evt.X)
			mob.Y = float64(evt.Y)
		}
	})
	g.EventBus.Subscribe((event.MobMove{}).Type(), func(e event.Event) {
		evt := e.(*event.MobMove)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.TargetX = float64(evt.X)
			mob.TargetY = float64(evt.Y)
			mob.TargetID = evt.TargetID
		}
	})

	// **** Request -> network send hooks.
	g.EventBus.Subscribe((request.Move{}).Type(), func(e event.Event) {
		// TODO: Send move request to server.
		// NOTE: We could do local interpolation here as well, so as to make the game feel more responsive in the event of lag.
	})
	g.EventBus.Subscribe((request.Leave{}).Type(), func(e event.Event) {
		// TODO: Send leave request to server.
	})
	g.EventBus.Subscribe((request.Construct{}).Type(), func(e event.Event) {
		// TODO: Send construct request to server.
	})
}

// Update updates the game state and processes events.
func (g *Game) Update() error {
	/*
		if g.localGame {
			g.ServerBus.ProcessEvents()
		}
	*/

	// Here is where we'd convert inputs, etc., into requests.

	// Update the thingz.
	g.EventBus.ProcessEvents()

	for _, mob := range g.Mobs {
		mob.Update(&g.State)
	}
	return nil
}

// Draw draws da game.
func (g *Game) Draw(screen *ebiten.Image) {
	for _, mob := range g.Mobs {
		g.DrawMob(screen, mob)
	}
}

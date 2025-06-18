package client

import (
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
	log *slog.Logger
	world.State
}

// Setup sets up our event and request hooks.
func (g *Game) Setup() {
	g.log = log.New("game", "client")
	g.EventBus = *event.NewBus("client")

	// **** Event -> local state change hooks.
	g.EventBus.Subscribe((event.MobSpawn{}).Type(), func(e event.Event) {
		evt := e.(*event.MobSpawn)
		mob := world.NewMob(evt.ID, evt.Owner, float64(evt.X), float64(evt.Y))
		g.Mobs.Add(mob)
		g.log.Debug("mob spawned", "id", evt.ID, "owner", evt.Owner, "x", evt.X, "y", evt.Y)
	})
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.X = float64(evt.X)
			mob.Y = float64(evt.Y)
			g.log.Debug("mob position updated", "id", evt.ID, "x", evt.X, "y", evt.Y)
		}
	})
	g.EventBus.Subscribe((event.MobMove{}).Type(), func(e event.Event) {
		evt := e.(*event.MobMove)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.TargetX = float64(evt.X)
			mob.TargetY = float64(evt.Y)
			mob.TargetID = evt.TargetID
			g.log.Debug("mob move requested", "id", evt.ID, "targetX", evt.X, "targetY", evt.Y, "targetID", evt.TargetID)
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

	// Update the thingz.
	g.EventBus.ProcessEvents()

	/*for _, mob := range g.Mobs {
		mob.Update(&g.State)
	}*/
	return nil
}

// Draw draws da game.
func (g *Game) Draw(screen *ebiten.Image) {
	for _, mob := range g.Mobs {
		g.DrawMob(screen, mob)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	// Return the original dimensions for now.
	return ow, oh
}

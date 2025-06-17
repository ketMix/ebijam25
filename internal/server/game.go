package server

import (
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/world"
)

// Player represents a player in the game gstance. It can be AI or a real hummus.
type Player struct {
	id int
	//conn *net.Connection
}

// Game represents a game instance. It is responsible for processgg the world.
type Game struct {
	world.State
	players []*Player
	//Resources []*Resource
}

// Setup sets up event subscriptions.
func (g *Game) Setup() {
	g.EventBus = *event.NewBus("server")
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.X = float64(evt.X)
			mob.Y = float64(evt.Y)
			// TODO: Periodically send mob position updates to players
		}
	})
}

// Update updates da world.
func (g *Game) Update() {
	/*for _, player := range g.Players {
		// Update player logic
	}*/

	for _, mob := range g.Mobs {
		g.UpdateMob(mob)
	}

	/*for _, resource := range g.Resources {
		// Update resource logic
	}*/
}

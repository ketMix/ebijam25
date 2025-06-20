package server

import (
	"fmt"
	"math/rand"

	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Setup sets up event subscriptions.
func (t *Table) Setup() {
	t.Seed = rand.Intn(1000000)                    // Random seed for world generation
	t.State.Continent = world.NewContinent(t.Seed) // Create a new continent with the seed and dimensions
	t.EventBus = *event.NewBus("table-" + fmt.Sprintf("%d", t.ID))
	t.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := t.Continent.Mobs.FindByID(evt.ID); mob != nil {
			t.Continent.MoveMob(mob, float64(evt.X), float64(evt.Y))
			// For now just send it, I guess.
			t.SendVisibleMobEvent(mob, e)
		}
	})

	// Subscribe to per-player messages. These are generated from the websockets listen loop for a given player connection.
	t.EventBus.Subscribe((PlayerMessage{}).Type(), func(e event.Event) {
		msg := e.(*PlayerMessage)
		switch evt := msg.msg.(type) {
		case *request.Move:
			if mob := t.Continent.Mobs.FindByID(msg.player.MobID); mob != nil {
				mob.TargetX = float64(evt.X)
				mob.TargetY = float64(evt.Y)
			} else {
				t.log.Warn("move request received but mob not found", "mobID", msg.player.MobID)
			}
		}
	})

	// Create a fake mob a distance away to test mob visibility.
	t.Continent.NewMob(2, t.mobID.Next(), 300, 300)
	t.Continent.NewMob(2, t.mobID.Next(), 200, 200)
}

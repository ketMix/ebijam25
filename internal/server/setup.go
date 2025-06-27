package server

import (
	"fmt"
	"math/rand/v2"

	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Setup sets up event subscriptions.
func (t *Table) Setup() {
	t.Seed = rand.Uint()
	t.State.Continent = world.NewContinent(t.Seed) // Create a new continent with the seed and dimensions
	t.EventBus = *event.NewBus("table-" + fmt.Sprintf("%d", t.ID))
	t.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := t.Continent.Mobs.FindByID(evt.ID); mob != nil {
			floatX := float64(evt.X) / world.FloatScale
			floatY := float64(evt.Y) / world.FloatScale
			t.Continent.MoveMob(mob, floatX, floatY)
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
				floatX := float64(evt.X) / world.FloatScale
				floatY := float64(evt.Y) / world.FloatScale
				mob.TargetX = floatX
				mob.TargetY = floatY
				e := &event.MobMove{
					ID:       mob.ID,
					X:        evt.X,
					Y:        evt.Y,
					TargetID: mob.TargetID,
				}
				t.SendVisibleMobEvent(mob, e)
			} else {
				t.log.Warn("move request received but mob not found", "mobID", msg.player.MobID)
			}
		}
	})

	// Create the director to manage the game contents
	t.director = NewDirector(t)
}

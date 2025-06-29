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
			t.Continent.MoveMob(mob, evt.X, evt.Y)
			// For now just send it, I guess.
			t.SendVisibleMobEvent(mob, e)

			// Check if we're intersecting with any other mobs.
			for _, other := range t.State.Continent.Mobs {
				if other.ID != mob.ID && mob.Intersects(other) {
					var baseDamage int
					switch mob.OuterKind {
					case world.SchlubKindPlayer:
						baseDamage = -5
					case world.SchlubKindMonk:
						baseDamage = -1 // Just overload -1 to mean conversion.
					case world.SchlubKindWarrior:
						baseDamage = 2
					case world.SchlubKindVagrant:
						baseDamage = 1
					}
					fmt.Println("Mob collision detected:", mob.ID, "with", other.ID, "base damage:", baseDamage)
					/*if baseDamage < 0 {
						// Convert schlubs from the other mob to this one.
						count := -baseDamage
						if count >= len(other.Schlubs) {
							count = len(other.Schlubs) - 1
						}
						if count > 0 {
							// Convert schlubs from the other mob to this one.
							var schlubIDs []int
							for i := 0; i < count; i++ {
								schlubIDs = append(schlubIDs, int(other.Schlubs[i]))
							}
							other.Schlubs = other.Schlubs[count:]
							mob.Schlubs = append(mob.Schlubs, other.Schlubs[:count-1]...)
							t.State.EventBus.Publish(&event.MobConvert{
								From: other.ID,
								To:   mob.ID,
								IDs:  schlubIDs,
							})
						}
					} else if baseDamage > 0 {
						if len(other.Schlubs) < baseDamage {
							baseDamage = len(other.Schlubs)
						}
						var schlubIDs []int
						for i := 0; i < baseDamage; i++ {
							schlubIDs = append(schlubIDs, int(other.Schlubs[i]))
						}

						// Deal damage to the other mob.
						t.State.EventBus.Publish(&event.MobDamage{
							ID:         other.ID,
							AttackerID: mob.ID,
							IDs:        schlubIDs,
						})
						// Remove the schlubs from the other mob.
						other.Schlubs = other.Schlubs[baseDamage:]
					}
					// Destroy the other mob if it has no schlubs left.
					if len(other.Schlubs) == 0 {
						t.State.EventBus.Publish(&event.MobDespawn{
							ID: other.ID,
						})
						// TODO: Maybe we should also check if the player has any mobs left.
					}*/
				}
			}
		}
	})
	t.EventBus.Subscribe((event.MobDamage{}).Type(), func(e event.Event) {
		evt := e.(*event.MobDamage)
		if mob := t.Continent.Mobs.FindByID(evt.ID); mob != nil {
			// If the mob is dead, remove it.
			if len(evt.IDs) == 0 {
				t.Continent.Mobs.Remove(mob)
				t.SendVisibleMobEvent(mob, &event.MobDespawn{
					ID: mob.ID,
				})
			} else {
				// Otherwise, just send the damage event.
				t.SendVisibleMobEvent(mob, evt)
			}
		} else {
			t.log.Warn("mob damage event received but mob not found", "id", evt.ID)
		}
	})
	t.EventBus.Subscribe((event.MobConvert{}).Type(), func(e event.Event) {
		evt := e.(*event.MobConvert)
		fromMob := t.Continent.Mobs.FindByID(evt.From)
		toMob := t.Continent.Mobs.FindByID(evt.To)
		if fromMob == nil || toMob == nil {
			t.log.Warn("mob convert event received but one or both mobs not found", "from", evt.From, "to", evt.To)
			return
		}
		// If the from mob is dead, remove it.
		if len(evt.IDs) == 0 {
			t.Continent.Mobs.Remove(fromMob)
			t.SendVisibleMobEvent(fromMob, &event.MobDespawn{
				ID: fromMob.ID,
			})
		} else {
			// Otherwise, convert the schlubs from the from mob to the to mob.
			for _, id := range evt.IDs {
				toMob.AddSchlub(world.SchlubID(id))
				fromMob.RemoveSchlub(world.SchlubID(id))
			}
			// Send the convert event to the to mob.
			t.SendVisibleMobEvent(toMob, &event.MobConvert{
				From: evt.From,
				To:   evt.To,
				IDs:  evt.IDs,
			})
		}
	})
	t.EventBus.Subscribe((event.MobCreate{}).Type(), func(e event.Event) {
		evt := e.(*event.MobCreate)
		// Just send it.
		if mob := t.Continent.Mobs.FindByID(evt.ID); mob != nil {
			// Add the new schlubs to the mob.
			for _, id := range evt.IDs {
				mob.AddSchlub(world.SchlubID(id))
			}
			// Send the spawn event to all players that can see the mob.
			t.SendVisibleMobEvent(mob, evt)
		} else {
			t.log.Warn("mob create event received but mob not found", "id", evt.ID)
		}
	})

	// Subscribe to per-player messages. These are generated from the websockets listen loop for a given player connection.
	t.EventBus.Subscribe((PlayerMessage{}).Type(), func(e event.Event) {
		msg := e.(*PlayerMessage)
		switch evt := msg.msg.(type) {
		case *request.Move:
			if mob := t.Continent.Mobs.FindByID(msg.player.MobID); mob != nil {
				mob.TargetX = evt.X
				mob.TargetY = evt.Y
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
		case *request.Formation:
			if mob := t.Continent.Mobs.FindByID(msg.player.MobID); mob != nil {
				// Eh... we're the arbiters of this.
				if mob.OuterKind == world.SchlubKindVagrant || mob.OuterKind == 0 {
					mob.OuterKind = world.SchlubKindMonk // Change the outer kind to monk
				} else if mob.OuterKind == world.SchlubKindMonk {
					mob.OuterKind = world.SchlubKindWarrior // Change the outer kind to warrior
				} else {
					mob.OuterKind = world.SchlubKindVagrant // Reset to vagrant
				}
				// Send that response.
				response := &event.MobFormation{
					ID:        mob.ID,
					OuterKind: int(mob.OuterKind),
				}
				t.SendVisibleMobEvent(mob, response)
			} else {
				t.log.Warn("formation request received but mob not found", "mobID", msg.player.MobID)
			}
		case *request.Construct:
			if evt.Caravan >= int(world.SchlubKindCaravanVagrant) && evt.Caravan <= int(world.SchlubKindCaravanWarrior) {
				if mob := t.Continent.Mobs.FindByID(msg.player.MobID); mob != nil {

					// Add a some schlubs.
					fam := t.FamilyID.NextSchlub()
					t.FamilyID = fam

					// Start with the player.
					fam.SetKindID(evt.Caravan) // Set the kind to Player
					mob.AddSchlub(fam)

					response := &event.MobCreate{
						ID:  msg.player.MobID,
						IDs: []int{int(fam)},
					}
					t.SendVisibleMobEvent(mob, response)
				}
			}
		}
	})

	// Create the director to manage the game contents
	t.director = NewDirector(t)
}

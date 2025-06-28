package server

import (
	"slices"

	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/world"
)

// UpdateMob updates the state of a mob in the game.
func (t *Table) UpdateMob(mob *world.Mob) {
	mob.Update(&t.State)
}

// RefreshVisibleMobs sends MobSpawn to players for mobs that are now visible and MobDespawn for mobs that are no longer visible.
func (t *Table) RefreshVisibleMobs(player *Player) {
	if player == nil {
		return
	}
	if mob := t.Continent.Mobs.FindByID(player.MobID); mob != nil {
		visibleMobs := t.Continent.Mobs.FindVisible(player.MobID)
		for _, visibleMob := range visibleMobs {
			if !slices.Contains(player.VisibleMobIDs, visibleMob.ID) {
				player.VisibleMobIDs = append(player.VisibleMobIDs, visibleMob.ID)
				// Send the new visible mob to the player
				t.log.Debug("new visible mob", "player", player.MobID, "mob", visibleMob.ID)
				t.SendMobTo(visibleMob, player)
			}
		}
		// Check for mobs that are no longer visible
		for i := len(player.VisibleMobIDs) - 1; i >= 0; i-- {
			if !slices.Contains(visibleMobs, t.Continent.Mobs.FindByID(player.VisibleMobIDs[i])) {
				t.HideMobFrom(player, t.Continent.Mobs.FindByID(player.VisibleMobIDs[i]))
				// Notify the player about the mob that is no longer visible
				t.log.Debug("mob no longer visible", "player", player.MobID, "mob", player.VisibleMobIDs[i])
				player.VisibleMobIDs = append(player.VisibleMobIDs[:i], player.VisibleMobIDs[i+1:]...)
			}
		}
	}
}

// HideMobFrom hides a mob from a player, so they can't see it anymore.
func (t *Table) HideMobFrom(player *Player, mob *world.Mob) {
	if player == nil || mob == nil {
		return
	}
	player.bus.Publish(&event.MobDespawn{
		ID: mob.ID,
	})
	t.log.Debug("mob hidden from player", "mob", mob.ID, "player", player.MobID)
}

// SendMobTo sends a mob to a player, so they can see it.
func (t *Table) SendMobTo(mob *world.Mob, player *Player) {
	if mob == nil || player == nil {
		return
	}

	var schlubs []int
	for _, s := range mob.Schlubs {
		schlubs = append(schlubs, int(s))
	}

	// Send the mob spawn event to the player.
	evt := &event.MobSpawn{
		ID:        mob.ID,
		Owner:     mob.OwnerID,
		X:         int(mob.X),
		Y:         int(mob.Y),
		Schlubs:   schlubs,
		OuterKind: int(mob.OuterKind),
	}

	player.bus.Publish(evt)

	t.log.Debug("mob sent to player", "mob", mob.ID, "player", player.MobID)
}

// SendVisibleMobEvent sends an event to all players that can see the mob.
func (t *Table) SendVisibleMobEvent(mob *world.Mob, evt event.Event) {
	for _, player := range t.players {
		if !slices.Contains(player.VisibleMobIDs, mob.ID) {
			continue
		}
		player.bus.Publish(evt)
	}
}

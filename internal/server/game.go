package server

import (
	"log/slog"
	"slices"

	"github.com/ketMix/ebijam25/internal/log"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Player represents a player in the game gstance. It can be AI or a real hummus.
type Player struct {
	world.Player // Just embed that shiz
	//conn *net.Connection
}

// Game represents a game instance. It is responsible for processgg the world.
type Game struct {
	world.State
	tickrate   int // Tickrate is the number of updates per second.
	log        *slog.Logger
	players    []*Player
	mobID      world.IDGenerator
	playerID   world.IDGenerator
	resourceID world.IDGenerator
	//Resources []*Resource
}

// Setup sets up event subscriptions.
func (g *Game) Setup() {
	g.log = log.New("game", "server")
	g.State.Tickrate = 5
	g.EventBus = *event.NewBus("server")
	g.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.Mobs.FindByID(evt.ID); mob != nil {
			mob.X = float64(evt.X)
			mob.Y = float64(evt.Y)
			// TODO: Periodically send mob position updates to players
			// For now just send it, I guess.
			//g.SendVisibleMobEvent(mob, e)
		}
	})

	// Subscribe to requests.
	g.EventBus.Subscribe((request.Join{}).Type(), func(e event.Event) {
		evt := e.(*request.Join)
		// for now just accept.
		player := world.NewPlayer(evt.Username, g.playerID.Next())
		g.players = append(g.players, &Player{
			Player: *player,
		})
		g.log.Debug("player joined", "username", evt.Username, "id", player.ID)
		g.EventBus.Publish(&event.MetaJoin{
			Username: evt.Username,
			ID:       player.ID,
		})
		// Create a new mob for the player.
		mob := world.NewMob(player.ID, g.mobID.Next(), 100, 100)
		g.players[len(g.players)-1].MobID = mob.ID
		g.Mobs.Add(mob)

		// Let the player know who they are.
		g.EventBus.Publish(&event.MetaWelcome{
			Username: evt.Username,
			ID:       player.ID,
			MobID:    mob.ID,
		})
		// TODO: Send all players to new player.
		// TODO: Send new player to all other players.
	})
	g.EventBus.Subscribe((request.Move{}).Type(), func(e event.Event) {
		evt := e.(*request.Move)
		if len(g.players) == 0 {
			g.log.Warn("move request received but no players exist")
			return
		}
		player := g.players[0] // For now, just use the first player.
		if mob := g.Mobs.FindByID(player.MobID); mob != nil {
			mob.TargetX = float64(evt.X)
			mob.TargetY = float64(evt.Y)
			g.SendVisibleMobEvent(mob, &event.MobMove{
				ID:       mob.ID,
				X:        int(mob.TargetX),
				Y:        int(mob.TargetY),
				TargetID: mob.TargetID,
			})
		} else {
			g.log.Warn("move request received but mob not found", "mobID", player.MobID)
		}
	})

	// Create a fake mob a distance away to test mob visibility.
	g.Mobs.Add(world.NewMob(2, g.mobID.Next(), 300, 300))
	g.Mobs.Add(world.NewMob(2, g.mobID.Next(), 200, 200))
	// For testing, let's add 100 schlubs.
	for range 100 {
		schlub := &world.Schlub{
			ID: g.mobID.Next(),
		}
		g.Mobs[len(g.Mobs)-1].Constituents = append(g.Mobs[len(g.Mobs)-1].Constituents, schlub)
	}
}

// Update updates da world.
func (g *Game) Update() {
	g.tickrate++
	if g.tickrate < g.State.Tickrate {
		return
	}
	g.tickrate = 0

	g.EventBus.ProcessEvents()

	for _, player := range g.players {
		g.RefreshVisibleMobs(player)
		// Send any unknown constituents to the player in delayed chunks.
		player.NextDelayedSend--
		if player.NextDelayedSend <= 0 {
			player.NextDelayedSend = g.tickrate + 1 // TODO: Make this an actual thought out value.
			if len(player.UnknownConstituents) > 0 {
				maxSend := 50
				if len(player.UnknownConstituents) < maxSend {
					maxSend = len(player.UnknownConstituents)
				}
				toSend := player.UnknownConstituents[:maxSend]

				var schlubCreate []event.SchlubCreate
				for _, constituentID := range toSend {
					schlubCreate = append(schlubCreate, event.SchlubCreate{
						ID: constituentID,
						// TODO: Other props.
					})
				}
				g.EventBus.Publish(&event.SchlubCreateList{
					Schlubs: schlubCreate,
				})
				player.UnknownConstituents = player.UnknownConstituents[maxSend:]
				player.KnownConstituents = append(player.KnownConstituents, toSend...)
			}
		}
	}

	for _, mob := range g.Mobs {
		g.UpdateMob(mob)
	}

	/*for _, resource := range g.Resources {
		// Update resource logic
	}*/
}

// RefreshVisibleMobs sends MobSpawn to players for mobs that are now visible and MobDespawn for mobs that are no longer visible.
func (g *Game) RefreshVisibleMobs(player *Player) {
	if player == nil {
		return
	}
	if mob := g.Mobs.FindByID(player.MobID); mob != nil {
		visibleMobs := g.Mobs.FindVisible(player.MobID)
		for _, visibleMob := range visibleMobs {
			if !slices.Contains(player.VisibleMobIDs, visibleMob.ID) {
				player.VisibleMobIDs = append(player.VisibleMobIDs, visibleMob.ID)
				// Send the new visible mob to the player
				g.log.Debug("new visible mob", "player", player.MobID, "mob", visibleMob.ID)
				g.SendMobTo(visibleMob, player)
			}
		}
		// Check for mobs that are no longer visible
		for i := len(player.VisibleMobIDs) - 1; i >= 0; i-- {
			if !slices.Contains(visibleMobs, g.Mobs.FindByID(player.VisibleMobIDs[i])) {
				g.HideMobFrom(player, g.Mobs.FindByID(player.VisibleMobIDs[i]))
				// Notify the player about the mob that is no longer visible
				g.log.Debug("mob no longer visible", "player", player.MobID, "mob", player.VisibleMobIDs[i])
				player.VisibleMobIDs = append(player.VisibleMobIDs[:i], player.VisibleMobIDs[i+1:]...)
			}
		}
	}
}

// SendMobTo sends a mob to a player, so they can see it.
func (g *Game) SendMobTo(mob *world.Mob, player *Player) {
	if mob == nil || player == nil {
		return
	}

	constituents := mob.ConstituentsToIDs()
	// These will be sent to the player over time.
	for _, constituent := range constituents {
		if !slices.Contains(player.KnownConstituents, constituent) {
			player.UnknownConstituents = append(player.UnknownConstituents, constituent)
		}
	}

	// Send the mob spawn event to the player.
	evt := &event.MobSpawn{
		ID:           mob.ID,
		Owner:        mob.OwnerID,
		X:            int(mob.X),
		Y:            int(mob.Y),
		Constituents: constituents, // Might as well send the constituent IDs as well. Probably not a big issue.
	}

	g.EventBus.Publish(evt)

	g.log.Debug("mob sent to player", "mob", mob.ID, "player", player.MobID)
}

// HideMobFrom hides a mob from a player, so they can't see it anymore.
func (g *Game) HideMobFrom(player *Player, mob *world.Mob) {
	if player == nil || mob == nil {
		return
	}
	g.EventBus.Publish(&event.MobDespawn{
		ID: mob.ID,
	})
	g.log.Debug("mob hidden from player", "mob", mob.ID, "player", player.MobID)
}

// SendVisibleMobEvent sends an event to all players that can see the mob.
func (g *Game) SendVisibleMobEvent(mob *world.Mob, evt event.Event) {
	for _, player := range g.players {
		if !slices.Contains(player.VisibleMobIDs, mob.ID) {
			continue
		}
		g.EventBus.Publish(evt)
	}
}

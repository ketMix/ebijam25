package server

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/log"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/world"
)

// Table is our table, dang. It's basically a full game structure.
type Table struct {
	world.State
	director       *Director
	log            *slog.Logger
	ID             world.ID
	open           bool
	running        bool
	players        []*Player
	playerID       world.IDGenerator // ID generator for players in this table
	playerAdd      chan *Player
	playerLeave    chan *Player
	playerMessages chan PlayerMessage // Channel for player messages
	mobID          world.IDGenerator
	resourceID     world.IDGenerator
	close          chan bool // Channel to signal table closure
}

const (
	debugSpawn = world.MaxSchlubsPerMob
)

// NewTable makes a new table, dang.
func NewTable(id world.ID) *Table {
	return &Table{
		ID:             id,
		log:            log.New("table", fmt.Sprintf("%d", id)),
		playerAdd:      make(chan *Player, 10),        // Buffered channel for player additions
		playerLeave:    make(chan *Player, 10),        // Buffered channel for player leave events
		close:          make(chan bool, 1),            // Buffered channel for closing the table
		playerMessages: make(chan PlayerMessage, 100), // Buffered channel for player messages
		open:           true,
		running:        true,
	}
}

// Loop is our table's loop that runs in a goroutine. It receives new players, player leaves, player messages, and runs the table's update function at a fixed tickrate.
func (t *Table) Loop() {
	t.Tickrate = 20                            // FIXME
	ticker := time.NewTicker(time.Second / 20) // 20 ticks per second
	// Process player additions
	for t.running {
		select {
		case <-t.close:
			t.log.Info("table closed")
			t.running = false
			// Boot all players from the table.
			for _, player := range t.players {
				player.conn.Close(websocket.StatusNormalClosure, "table closed")
			}
			// FIXME: This should just get players into a new table.
			return // Exit the loop if the table is closed
		case msg := <-t.playerMessages:
			t.EventBus.Publish(&msg) // Publish the message to the event bus
		case player := <-t.playerAdd:
			t.AddPlayer(player)
			// Send a welcome message to the new player.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			// Create a new mob for the player.
			x, y := t.director.GetSpawnPosition()
			mob := t.Continent.NewMob(player.ID, t.mobID.Next(), x, y)
			player.MobID = mob.ID // Assign the mob ID to the player

			// Add a some schlubs.
			fam := t.FamilyID.NextFamily()
			t.FamilyID = fam

			// Start with the player.
			fam = fam.NextSchlub()
			fam.SetKindID(int(world.SchlubKindPlayer)) // Set the kind to Player
			mob.AddSchlub(fam)

			// Perhaps a little unfair (due to some people getting' ROBBED), but let's give a few random schlubs to the player.
			for range 4 {
				if t.Continent.Fate.NumGen.Intn(100) < 50 { // 50% chance to add a random schlub
					if t.Continent.Fate.NumGen.Intn(100) < 50 { // 50% chance for it to be from a diff. fam.
						fam = fam.NextFamily()
					} else {
						fam = fam.NextSchlub() // Just get the next schlub in the same family
					}
					fam.SetKindID(int(world.SchlubKindVagrant)) // Set the kind to Vagrant
					mob.AddSchlub(fam)
				}
			}

			/*kindId := int(world.SchlubKindVagrant)
			for range debugSpawn {
				fam = fam.NextSchlub()
				fam.SetKindID(kindId)
				mob.AddSchlub(fam)
				kindId++
				if kindId > int(world.SchlubKindWarrior) {
					kindId = int(world.SchlubKindVagrant)
				}
			}*/

			welcome, _ := message.Encode(&event.MetaWelcome{
				Username: player.Username,
				ID:       player.ID,
				Color:    player.Color,
				MobID:    mob.ID,
				Seed:     t.Seed,
				Rate:     t.State.Tickrate,
			})
			player.conn.Write(ctx, websocket.MessageText, welcome)
			// Also send a join event to all players.
			for _, p := range t.players {
				if p.ID != player.ID { // Don't send to the new player
					joinEvent, _ := message.Encode(&event.MetaJoin{
						Username: player.Username,
						Color:    player.Color,
						ID:       player.ID,
					})
					player.conn.Write(ctx, websocket.MessageText, joinEvent)
				}
			}
			// Mark the table as closed in there are 15+ players.
			if len(t.players) >= 15 {
				t.open = false // Close the table for new players
			}
		case player := <-t.playerLeave:
			// Handle player leaving the table
			for i, p := range t.players {
				if p.ID == player.ID {
					t.players = append(t.players[:i], t.players[i+1:]...) // Remove player from the slice
					break
				}
			}
			for _, p := range t.players {
				leaveEvent, _ := message.Encode(&event.MetaLeave{
					ID: player.ID,
				})
				p.conn.Write(context.Background(), websocket.MessageText, leaveEvent) // Notify other players about the player leaving
			}
			// TODO: Notify other players about the player leaving
			for _, mob := range t.Continent.Mobs {
				if mob.OwnerID == player.ID {
					for _, p := range t.players {
						t.Continent.RemoveMob(mob) // Remove the mob associated with the player
						despawnEvent, _ := message.Encode(&event.MobDespawn{
							ID: mob.ID,
						})
						p.conn.Write(context.Background(), websocket.MessageText, despawnEvent)
					}
				}
			}
		case <-ticker.C:
			// process da world, my final message
			t.Update()
		}
	}
}

// Update updates da world.
func (t *Table) Update() {
	t.EventBus.ProcessEvents()

	for _, player := range t.players {
		t.RefreshVisibleMobs(player)
		// Also periodically refresh all player info.
		player.lastRefresh++
		if player.lastRefresh > 30 { // Refresh every 30 ticks
			player.lastRefresh = 0
			for _, p := range t.players {
				if mob := t.Continent.Mobs.FindByID(p.MobID); mob != nil {
					refreshEvent, _ := message.Encode(&event.MetaRefresh{
						ID:    p.ID,
						Count: len(mob.Schlubs),
					})
					player.conn.Write(context.Background(), websocket.MessageText, refreshEvent)
				}
			}
		}

		player.bus.ProcessEvents()
	}
	t.director.Update()
	t.UpdateContinent()
}

// AddPlayer adds a player, hooks up buses, and starts a goroutine to handle player messages.
func (t *Table) AddPlayer(player *Player) {
	player.ID = t.playerID.Next() // Assign a new ID to the player
	t.players = append(t.players, player)
	// Hook up that busy ;) (this writes all events received on the bus to the player's websocket connection)
	player.bus.SubscribePrefix("", func(e event.Event) {
		data, err := message.Encode(e)
		if err != nil {
			fmt.Println("error encoding message:", err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		err = player.conn.Write(ctx, websocket.MessageText, data)
		if err != nil {
			fmt.Println("error writing to player connection:", err)
			return
		}
	})

	// It's a bit crap, but we need to spawn a new goroutine for each player.
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			player.cancel = cancel // Store the cancel function in the player struct
			kind, data, err := player.conn.Read(ctx)
			if err != nil {
				fmt.Println("error reading from player connection:", err)
				break
			}
			if kind != websocket.MessageText {
				continue
			}

			msg, err := message.Decode(data)
			if err != nil {
				fmt.Println("error decoding message:", err)
				break
			}
			t.playerMessages <- PlayerMessage{
				player: player,
				msg:    msg,
			}
		}

		// Yeet the table if there are no players left and we closed it.
		if len(t.players) == 0 && !t.open {
			t.close <- true // Signal the table to close
			t.log.Info("table closed due to no players")
			return
		}

		t.playerLeave <- player // Notify the table that the player is leaving
		player.conn.Close(websocket.StatusNormalClosure, "bai")
	}()
}

// Tables is our tables.
type Tables struct {
	tables []*Table
	idGen  world.IDGenerator
}

// AcquireOpenTable either creates a new open table and spawns a goroutine to handle it or returns an existing one.
func (t *Tables) AcquireOpenTable() *Table {
	for _, table := range t.tables {
		if table.open {
			return table
		}
	}
	newTable := NewTable(t.idGen.Next())
	newTable.Setup()
	t.tables = append(t.tables, newTable)
	// Spin it up...
	go newTable.Loop()
	return newTable
}

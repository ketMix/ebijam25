package world

import (
	"github.com/ketMix/ebijam25/internal/message/event"
)

// State represents the current state of the game world.
type State struct {
	Tickrate int // The current tick rate of the world.
	EventBus event.Bus
	Mobs     Mobs // Collection of mobs in the world
	PlayerID ID   // The ID of the local player
	MobID    ID   // The ID of the local player's mob
}

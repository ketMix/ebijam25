package world

import (
	"github.com/ketMix/ebijam25/internal/message/event"
)

// State represents the current state of the game world.
type State struct {
	EventBus event.Bus
	Mobs     Mobs // Collection of mobs in the world
}

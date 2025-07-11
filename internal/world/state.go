package world

import (
	"image/color"

	"github.com/ketMix/ebijam25/internal/message/event"
)

// State represents the current state of the game world.
type State struct {
	Seed      uint // The seed used for world generation.
	Tickrate  int  // The current tick rate of the world.
	EventBus  event.Bus
	Continent *Continent  // The current continent of the game world.
	PlayerID  ID          // The ID of the local player
	MobID     ID          // The ID of the local player's mob
	Color     color.NRGBA // The color of the local player in NRGBA format.
	FamilyID  SchlubID    // Family schlub generator
}

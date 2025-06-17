package event

import "github.com/ketMix/ebijam25/internal/message"

// MobMerge represents an event where one mob is merged into another.
type MobMerge struct {
	From int `json:"from"` // ID of the mob being merged
	To   int `json:"to"`   // ID of the mob being merged into
}

// Type returns the type of the MobMerge event.
func (m MobMerge) Type() string {
	return "mob-merge"
}

// MobSplit represents an event where a mob is split into another mob.
type MobSplit struct {
	ID int `json:"from"` // ID of the mob being split
	// FIXME: This will not be cloned properly in Decode.
	Schlubs []int `json:"schlubs"` // IDs of schlubs being split
}

// Type returns the type of the MobSplit event.
func (m MobSplit) Type() string {
	return "mob-split"
}

// MobMove represents an event where a mob begins to move to a new position.
type MobMove struct {
	ID int `json:"id"` // ID of the mob moving
	X  int `json:"x"`
	Y  int `json:"y"`
}

// Type returns the type of the MobMove event.
func (m MobMove) Type() string {
	return "mob-move"
}

// MobSpawn represents an event where a new mob is spawned at a specific location. It is required that schlubs are created prior to this event.
type MobSpawn struct {
	ID int `json:"id"` // ID of the spawned mob
	X  int `json:"x"`
	Y  int `json:"y"`
	// FIXME: This will not be cloned properly in Decode.
	Schlubs []int `json:"schlubs"` // IDs of schlubs associated with the mob
}

// Type returns the type of the MobSpawn event.
func (m MobSpawn) Type() string {
	return "mob-spawn"
}

func init() {
	message.Register(MobMerge{})
	message.Register(MobSplit{})
	message.Register(MobMove{})
	message.Register(MobSpawn{})
}

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
	ID       int `json:"id"`        // ID of the mob moving
	TargetID int `json:"target_id"` // ID of the target mob (if applicable)
	X        int `json:"x"`
	Y        int `json:"y"`
}

// Type returns the type of the MobMove event.
func (m MobMove) Type() string {
	return "mob-move"
}

// MobPosition represents the position of a mob on the map.
type MobPosition struct {
	ID int `json:"id"` // ID of the mob
	X  int `json:"x"`
	Y  int `json:"y"`
}

// Type returns the type of the MobPosition event.
func (m MobPosition) Type() string {
	return "mob-position"
}

// MobSpawn represents an event where a new mob is spawned at a specific location. It is required that schlubs are created prior to this event.
type MobSpawn struct {
	ID    int `json:"id"`    // ID of the spawned mob
	Owner int `json:"owner"` // ID of the owner (player) of the mob
	X     int `json:"x"`
	Y     int `json:"y"`
	// FIXME: This will not be cloned properly in Decode.
	Schlubs   []int    `json:"schlubs"`             // IDs of schlubs associated with the mob
	Formation []string `json:"formation,omitempty"` // Optional formation order of schlubs
}

// Type returns the type of the MobSpawn event.
func (m MobSpawn) Type() string {
	return "mob-spawn"
}

// MobDespawn represents an event where a mob is despawned.
type MobDespawn struct {
	ID int `json:"id"` // ID of the mob being despawned
}

// Type returns the type of the MobDespawn event.
func (m MobDespawn) Type() string {
	return "mob-despawn"
}

// MobDamage represents an even where a mob taks some damage, yo.
type MobDamage struct {
	ID     int `json:"id"`     // ID of the mob taking damage
	Amount int `json:"amount"` // Amount of damage taken, should be number of schlubs killed
}

// Type is a friggin' method that returns a dumbass string to uniquely identify the type, WOW.
func (m MobDamage) Type() string {
	return "mob-damage"
}

// MobConvert represents an event where schlubs from one mob are converted to be in another mob.
type MobConvert struct {
	From   int `json:"from"`   // ID of the mob being converted from.
	To     int `json:"to"`     // ID of the mob being converted into.
	Amount int `json:"amount"` // Amount of schlubs being converted.
}

// Type do the thing you expect it be doin, truly WOW.
func (m MobConvert) Type() string {
	return "mob-convert"
}

// MobFormation is a response to request-formation.
type MobFormation struct {
	ID        int      `json:"id"` // ID of the mob forming.
	Formation []string `json:"formation"`
}

// Type is a type.
func (m MobFormation) Type() string {
	return "mob-formation"
}

func init() {
	message.Register(&MobMerge{})
	message.Register(&MobSplit{})
	message.Register(&MobMove{})
	message.Register(&MobPosition{})
	message.Register(&MobSpawn{})
	message.Register(&MobDespawn{})
	message.Register(&MobDamage{})
	message.Register(&MobConvert{})
	message.Register(&MobFormation{})
}

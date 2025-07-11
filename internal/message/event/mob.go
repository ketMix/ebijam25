package event

import (
	"github.com/ketMix/ebijam25/internal/message"
)

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
	ID       int     `json:"id"`        // ID of the mob moving
	TargetID int     `json:"target_id"` // ID of the target mob (if applicable)
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
}

// Type returns the type of the MobMove event.
func (m MobMove) Type() string {
	return "mob-move"
}

// MobPosition represents the position of a mob on the map.
type MobPosition struct {
	ID int     `json:"id"` // ID of the mob
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

// Type returns the type of the MobPosition event.
func (m MobPosition) Type() string {
	return "mob-position"
}

// MobSpawn represents an event where a new mob is spawned at a specific location. It is required that schlubs are created prior to this event.
type MobSpawn struct {
	ID    int     `json:"id"`    // ID of the spawned mob
	Owner int     `json:"owner"` // ID of the owner (player) of the mob
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	// FIXME: This will not be cloned properly in Decode.
	Schlubs   []int `json:"schlubs"`         // IDs of schlubs associated with the mob
	OuterKind int   `json:"outer,omitempty"` // Optional outer kind of the mob, used for formation
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

// MobCreate adds the given schlubs to a mob.
type MobCreate struct {
	ID  int   `json:"id"`
	IDs []int `json:"schlubs"`
}

// Type returns garbage.
func (m MobCreate) Type() string {
	return "mob-create"
}

// MobDamage represents an even where a mob taks some damage, yo.
type MobDamage struct {
	ID         int   `json:"id"`          // ID of the mob taking damage
	AttackerID int   `json:"attacker_id"` // ID of the mob that dealt the damage
	IDs        []int `json:"schlubs"`     // IDs of the mobs being killed.
}

// Type is a friggin' method that returns a dumbass string to uniquely identify the type, WOW.
func (m MobDamage) Type() string {
	return "mob-damage"
}

// MobConvert represents an event where schlubs from one mob are converted to be in another mob.
type MobConvert struct {
	From int   `json:"from"`    // ID of the mob being converted from.
	To   int   `json:"to"`      // ID of the mob being converted into.
	IDs  []int `json:"schlubs"` // IDs of the mobs being pulled.
}

// Type do the thing you expect it be doin, truly WOW.
func (m MobConvert) Type() string {
	return "mob-convert"
}

// MobFormation is a response to request-formation.
type MobFormation struct {
	ID        int `json:"id"` // ID of the mob forming.
	OuterKind int `json:"outer"`
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
	message.Register(&MobCreate{})
}

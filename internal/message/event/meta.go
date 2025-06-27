package event

import "github.com/ketMix/ebijam25/internal/message"

// MetaJoin represents an event where a player joins the game with a username and unique ID.
type MetaJoin struct {
	Username string `json:"username"`
	ID       int    `json:"id"`
}

// Type returns the type of the MetaJoin event.
func (m MetaJoin) Type() string {
	return "meta-join"
}

// MetaWelcome represents a welcome event for a player joining the game. It is the counterpart to MetaJoin.
type MetaWelcome struct {
	Username string `json:"username"`
	ID       int    `json:"id"`
	MobID    int    `json:"mobId"` // ID of the mob associated with the player
	Seed     uint   `json:"seed"`  // Seed for this game's continent generation
	Rate     int    `json:"rate"`  // Tick
}

// Type returns the type of the MetaWelcome event.
func (m MetaWelcome) Type() string {
	return "meta-welcome"
}

// MetaLeave represents an event where a player leaves the game.
type MetaLeave struct {
	ID int `json:"id"`
}

// Type returns the type of the MetaLeave event.
func (m MetaLeave) Type() string {
	return "meta-leave"
}

func init() {
	message.Register(&MetaJoin{})
	message.Register(&MetaWelcome{})
	message.Register(&MetaLeave{})
}

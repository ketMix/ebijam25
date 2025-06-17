package event

// MetaJoin represents an event where a player joins the game with a username and unique ID.
type MetaJoin struct {
	Username string `json:"username"`
	ID       int    `json:"id"`
}

// Type returns the type of the MetaJoin event.
func (m MetaJoin) Type() string {
	return "meta-join"
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
	Register(MetaJoin{})
	Register(MetaLeave{})
}

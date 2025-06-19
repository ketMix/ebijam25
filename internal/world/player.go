package world

// Player represents a player in the game world.
type Player struct {
	Username            string // Player's username
	ID                  int    // Player's unique ID
	MobID               ID     // Player's unique mob ID
	VisibleMobIDs       []ID   // List of mobs visible to the player
	KnownConstituents   []ID   // List of known constituents to the player
	UnknownConstituents []ID   // List of unknown constituents to the player -- these are SPREAD.
	NextDelayedSend     int    // Next time to send delayed messages like unknown constituents
}

// NewPlayer makes a new player, wow.
func NewPlayer(user string, id ID) *Player {
	return &Player{
		Username: user,
		ID:       id,
	}
}

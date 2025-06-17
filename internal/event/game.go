package event

// GameWin represents an event where a player wins the game.
type GameWin struct {
	ID int `json:"id"` // Player ID who won the game
	// TODO: Victory type?
}

// Type returns the type of the GameWin event.
func (g GameWin) Type() string {
	return "game-win"
}

// GameOver represents an event where a player loses the game.
type GameOver struct {
	ID int `json:"id"` // Player ID who lost the game
}

// Type returns the type of the GameOver event.
func (g GameOver) Type() string {
	return "game-over"
}

func init() {
	Register(GameWin{})
	Register(GameOver{})
}

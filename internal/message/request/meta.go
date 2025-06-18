package request

import (
	"github.com/ketMix/ebijam25/internal/message"
)

// Join represents a request to join the game with a username.
type Join struct {
	Username string `json:"username"`
}

// Type returns the type of the Join request.
func (j Join) Type() string {
	return "request-join"
}

// Leave represents a request to leave the game.
type Leave struct {
}

// Type returns the type of the Leave request.
func (l Leave) Type() string {
	return "request-leave"
}

func init() {
	message.Register(Join{})
	message.Register(Leave{})
}

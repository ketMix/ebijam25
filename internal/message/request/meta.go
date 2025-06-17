package request

import (
	"github.com/ketMix/ebijam25/internal/message"
)

// Leave represents a request to leave the game.
type Leave struct {
}

// Type returns the type of the Leave request.
func (l Leave) Type() string {
	return "request-leave"
}

func init() {
	message.Register(Leave{})
}

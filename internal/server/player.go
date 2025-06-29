package server

import (
	"context"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/world"
)

// Player represents a player in the game gstance. It can be AI or a real hummus.
type Player struct {
	world.Player // Just embed that shiz
	bus          event.Bus
	conn         *websocket.Conn
	cancel       context.CancelFunc
	lastRefresh  int
}

// PlayerMessage is a wrapper around messages to attach a player to it. This is used to ensure that messages received by a connection are mapped to their appropriate player.
type PlayerMessage struct {
	player *Player
	msg    message.MessageI
}

// Type provides conformance to message.MessageI, yo.
func (m PlayerMessage) Type() string {
	return "player-message"
}

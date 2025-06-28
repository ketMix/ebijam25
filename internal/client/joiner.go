package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/event"
)

// Joiner is a badly named struct that handles joining a server.
type Joiner struct {
	conn     *websocket.Conn
	cancel   context.CancelFunc
	canceled chan bool
}

// Send does what you'd expect.
func (j *Joiner) Send(msg message.MessageI) {
	if j.conn == nil {
		panic("joiner connection is nil, cannot send message")
	}
	data, err := message.Encode(msg)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = j.conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		panic(err)
	}
}

// Join joins the given host and publishes any received messages to the provided event bus.
func (j *Joiner) Join(secure bool, host string, bus *event.Bus) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if secure {
		host = "wss://" + host
	} else {
		host = "ws://" + host
	}

	c, _, err := websocket.Dial(ctx, host, nil)
	if err != nil {
		panic(err)
	}

	j.conn = c

	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			j.cancel = cancel

			kind, data, err := c.Read(ctx)
			if err != nil {
				panic(err)
			}
			if kind != websocket.MessageText {
				continue
			}

			msg, err := message.Decode(data)
			if err != nil {
				println("error decoding message:", err.Error())
				break
			} else {
				bus.Publish(msg)
			}
		}

		c.Close(websocket.StatusNormalClosure, "bai bai")
		j.conn = nil
		j.canceled <- true
	}()
}

// Stoppe stoppes.
func (j *Joiner) Stoppe() {
	if j.cancel != nil {
		j.cancel()
	}
}

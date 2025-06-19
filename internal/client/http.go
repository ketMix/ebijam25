package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/request"
)

type Joiner struct {
	cancel   context.CancelFunc
	canceled chan bool
}

func (j *Joiner) Join(host string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://"+host, nil)
	if err != nil {
		panic(err)
	}

	data, err := message.Encode(&request.Join{
		Username: "Throbbing John",
	})

	if err != nil {
		panic(err)
	}

	err = c.Write(ctx, websocket.MessageText, data)
	if err != nil {
		panic(err)
	}

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
				println("decoded message:", msg.Type())
			}
		}

		c.Close(websocket.StatusNormalClosure, "bai bai")
		j.canceled <- true
	}()
}

func (j *Joiner) Stoppe() {
	if j.cancel != nil {
		j.cancel()
	}
}

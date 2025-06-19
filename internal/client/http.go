package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/request"
)

type Joiner struct {
}

func (j *Joiner) Join(host string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://"+host, nil)
	if err != nil {
		panic(err)
	}
	defer c.CloseNow()

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

	c.Close(websocket.StatusNormalClosure, "bai bai")
}

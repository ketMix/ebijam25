package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
)

type Listener struct {
	cancel context.CancelFunc
}

func (l *Listener) Listen(port int) {
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, err := websocket.Accept(w, r, nil)
				if err != nil {
					panic(err)
				}
				defer c.CloseNow()

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()

				kind, data, err := c.Read(ctx)
				if err != nil {
					panic(err)
				}
				fmt.Println("got", kind, data, string(data))
				msg, err := message.Decode(data)
				if err != nil {
					fmt.Println("error decoding message:", err)
				} else {
					fmt.Println("decoded message:", msg.Type())
				}

				c.Close(websocket.StatusNormalClosure, "bai")
			}),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to start server on port %d: %v", port, err))
		}
	}()
}

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/event"
)

type Listener struct {
	cancel   context.CancelFunc
	canceled chan bool
}

func (l *Listener) Listen(port int) {
	l.canceled = make(chan bool, 1)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c, err := websocket.Accept(w, r, nil)
				if err != nil {
					panic(err)
				}

				for {
					ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
					l.cancel = cancel

					_, data, err := c.Read(ctx)
					if err != nil {
						panic(err)
					}
					msg, err := message.Decode(data)
					if err != nil {
						fmt.Println("error decoding message:", err)
						break
					} else {
						fmt.Println("decoded message:", msg.Type())
					}
					if msg.Type() == "request-join" {
						welcome, err := message.Encode(&event.MetaWelcome{
							Username: "Throbbing John",
							ID:       1,
							MobID:    1, // Example mob ID
						})
						if err != nil {
							fmt.Println("error encoding welcome message:", err)
							break
						}
						c.Write(ctx, websocket.MessageText, welcome)
					}
				}

				c.Close(websocket.StatusNormalClosure, "bai")
				l.canceled <- true
			}),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to start server on port %d: %v", port, err))
		}
	}()
}

func (l *Listener) Stoppe() {
	if l.cancel != nil {
		l.cancel()
	}
}

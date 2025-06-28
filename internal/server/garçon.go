package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/ketMix/ebijam25/internal/message"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/world"
)

// Garçon governs getting clients to their game.
type Garçon struct {
	canceled chan bool
	tables   Tables
}

func (g *Garçon) Serve(port int, shouldGoroutine bool) {
	g.canceled = make(chan bool, 1)
	if shouldGoroutine {
		go g.listen(port)
	} else {
		g.listen(port)
	}
}

func (g *Garçon) listen(port int) {
	err := http.ListenAndServe(fmt.Sprintf(":%d", port),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Upgrade") != "websocket" {
				if r.URL.Path == "/" || r.URL.Path == "/index.html" {
					http.ServeFile(w, r, "web/index.html")
				} else if r.URL.Path == "/wasm_exec.js" {
					http.ServeFile(w, r, "web/wasm_exec.js")
				} else if r.URL.Path == "/ebijam25.wasm" {
					http.ServeFile(w, r, "web/ebijam25.wasm")
				} else {
					http.NotFound(w, r)
				}
				return
			}

			c, err := websocket.Accept(w, r, nil)
			if err != nil {
				panic(err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
			defer cancel()

			_, data, err := c.Read(ctx)
			if err != nil {
				panic(err)
			}
			msg, err := message.Decode(data)
			if err != nil {
				fmt.Println("error decoding message:", err)
				return
			}
			if msg.Type() == "request-join" {
				msg := msg.(*request.Join)
				// Let's get a table for 'em.
				table := g.tables.AcquireOpenTable()
				if table == nil {
					c.Close(websocket.StatusTryAgainLater, "no open tables")
					return
				}
				player := world.NewPlayer(msg.Username, -1)
				table.playerAdd <- &Player{
					Player: *player,
					bus:    *event.NewBus("player-" + player.Username),
					conn:   c,
				}
			}
		}),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start server on port %d: %v", port, err))
	}
}

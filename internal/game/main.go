package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/client"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/server"
)

type Game struct {
	Managers  Managers
	client    client.Game
	localGame bool
	garçon    server.Garçon
}

func NewGame(localGame bool) *Game {
	g := &Game{
		localGame: localGame,
	}

	g.client.Setup()
	g.client.EventBus.NoQueue = true

	if localGame {
		// Subscribe to our own requests to automatically network send them.
		g.client.EventBus.SubscribePrefix("request-", func(e event.Event) {
			g.client.Send(e)
		})

		// Spin up our garçon and join it.
		g.garçon.Serve(8080, true)
		g.client.Join("localhost:8080", &g.client.EventBus)

		// Send our join request with our name.
		g.client.EventBus.Publish(&request.Join{
			Username: "Player1",
		})
	}

	return g
}

func (g *Game) Update() error {
	g.Managers.Update()
	if err := g.client.Update(); err != nil {
		return err
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{32, 0, 32, 255})
	g.Managers.Draw(screen)
	g.client.Draw(screen)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	g.client.Layout(ow, oh)
	return ow, oh
}

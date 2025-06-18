package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/client"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/server"
)

type Game struct {
	Managers  Managers
	client    client.Game
	localGame bool
	server    server.Game
}

func NewGame(localGame bool) *Game {
	g := &Game{
		localGame: localGame,
	}

	g.client.Setup()

	if localGame {
		g.server.Setup()
		g.server.EventBus.Pipe(&g.client.EventBus, []string{"mob-", "schlub-", "meta-"})
		g.client.EventBus.Pipe(&g.server.EventBus, []string{"request-"})

		g.client.EventBus.Publish(&request.Join{
			Username: "Player1",
		})
	}

	return g
}

func (g *Game) Update() error {
	g.Managers.Update()
	if g.localGame {
		g.server.Update()
	}
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

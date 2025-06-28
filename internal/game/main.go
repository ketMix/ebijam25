package game

import (
	"image/color"
	"math/rand"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/client"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/message/request"
	"github.com/ketMix/ebijam25/internal/server"
	"github.com/kettek/rebui"
	"github.com/kettek/rebui/widgets"
)

type Game struct {
	Managers  Managers
	client    client.Game
	localGame bool
	garçon    server.Garçon
	layout    rebui.Layout
}

func NewGame(localGame bool) *Game {
	g := &Game{
		localGame: localGame,
	}

	g.client.Setup()
	g.client.EventBus.NoQueue = true

	// Subscribe to our own requests to automatically network send them.
	g.client.EventBus.SubscribePrefix("request-", func(e event.Event) {
		g.client.Send(e)
	})

	if localGame {
		// Spin up our garçon and join it.
		g.garçon.Serve(9099, true)
		g.client.Join(false, "localhost:9099", &g.client.EventBus)
	} else {
		g.client.Join(true, "schlubs.gamu.group", &g.client.EventBus)
	}

	// Set up some layout.
	var clr color.NRGBA
	var colorNode *rebui.Node
	node := g.layout.AddNode(rebui.Node{
		Type:            "TextInput",
		ID:              "name",
		Width:           "50%",
		Height:          "30",
		X:               "50%",
		Y:               "50%",
		OriginX:         "-50%",
		OriginY:         "-50%",
		ForegroundColor: "white",
		BackgroundColor: "black",
		BorderColor:     "white",
		VerticalAlign:   rebui.AlignMiddle,
		HorizontalAlign: rebui.AlignCenter,
		FocusIndex:      1,
	})
	node.Widget.(*widgets.TextInput).OnSubmit = func(s string) {
		g.client.EventBus.Publish(&request.Join{
			Username: s,
			Color:    clr,
		})
		g.client.Joined = true
		g.layout.RemoveNode(node)
		g.layout.RemoveNode(colorNode)
	}
	colorNode = g.layout.AddNode(rebui.Node{
		Type:            "TextInput",
		Width:           "20%",
		Height:          "30",
		X:               "after name",
		Y:               "at name",
		ForegroundColor: "white",
		BackgroundColor: "black",
		BorderColor:     "white",
		VerticalAlign:   rebui.AlignMiddle,
		HorizontalAlign: rebui.AlignCenter,
		FocusIndex:      1,
	})
	// Randomize the initial color.
	clr.R = uint8(100 + rand.Intn(155))
	clr.G = uint8(100 + rand.Intn(155))
	clr.B = uint8(100 + rand.Intn(155))
	clr.A = 255
	colorNode.Widget.(*widgets.TextInput).AssignText("#" + strconv.FormatUint(uint64(clr.R), 16) +
		strconv.FormatUint(uint64(clr.G), 16) +
		strconv.FormatUint(uint64(clr.B), 16))
	colorNode.Widget.(*widgets.TextInput).OnChange = func(s string) {
		clr = stringToColor(s, color.NRGBA{255, 255, 255, 255})
		clr.A = 255 // Ensure alpha is always 255.
	}

	return g
}

func (g *Game) Update() error {
	g.Managers.Update()
	if err := g.client.Update(); err != nil {
		return err
	}
	g.layout.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{32, 0, 32, 255})
	g.Managers.Draw(screen)
	g.client.Draw(screen)
	g.layout.Draw(screen)
}

func (g *Game) Layout(ow, oh int) (int, int) {
	g.layout.Layout(float64(ow), float64(oh))
	g.client.Layout(ow, oh)
	return ow, oh
}

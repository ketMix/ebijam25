package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kettek/rebui"
	_ "github.com/kettek/rebui/defaults/font"
	"github.com/kettek/rebui/widgets"
)

type Debug struct {
	layout   rebui.Layout
	leftNode *rebui.Node
}

func (d *Debug) Setup() {
	d.leftNode = d.layout.AddNode(rebui.Node{
		Type: "Text",
	})
}

func (d *Debug) setLeftText(text string) {
	d.leftNode.Widget.(*widgets.Text).AssignText(text)
}

func (d *Debug) Update() {
	d.layout.Update()
}

func (d *Debug) Draw(screen *ebiten.Image) {
	d.layout.Draw(screen)
}

package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kettek/rebui"
	"github.com/kettek/rebui/widgets"
)

type Resource struct {
	id        int
	x, y      float64
	food      int
	debugNode *rebui.Node
}

var resourceIdCounter = 0

func nextResourceId() int {
	resourceIdCounter++
	return resourceIdCounter
}

func NewResource(x, y float64, food int) *Resource {
	node := gLayout.AddNode(rebui.Node{
		Type:            "Text",
		Width:           "128",
		Height:          "32",
		X:               fmt.Sprintf("%f", x-float64(food)),
		Y:               fmt.Sprintf("%f", y-float64(food)),
		Text:            fmt.Sprintf("%d", food),
		ForegroundColor: "white",
	})
	node.Widget.(*widgets.Text).AssignBorderColor(nil)
	return &Resource{
		id:        nextResourceId(),
		x:         x,
		y:         y,
		food:      food,
		debugNode: node,
	}
}

func (r *Resource) Draw(screen *ebiten.Image) {
	size := float64(r.food)
	vector.DrawFilledRect(screen, float32(r.x-size/2), float32(r.y-size/2), float32(size), float32(size), color.NRGBA{255, 255, 0, 255}, false)
}

func (r *Resource) Deplete(amount int) {
	r.food -= amount
	if r.food < 0 {
		r.food = 0
		eventBus.Publish(&EventResourceDepleted{
			id: r.id,
		})
	}
	r.debugNode.Widget.(*widgets.Text).AssignText(fmt.Sprintf("%d", r.food))
}

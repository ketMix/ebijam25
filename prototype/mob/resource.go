package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Resource struct {
	id   int
	x, y float64
	food int
}

var resourceIdCounter = 0

func nextResourceId() int {
	resourceIdCounter++
	return resourceIdCounter
}

func NewResource(x, y float64, food int) *Resource {
	return &Resource{
		id:   nextResourceId(),
		x:    x,
		y:    y,
		food: food,
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
}

package client

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/message/event"
	"github.com/ketMix/ebijam25/internal/world"
)

type Game struct {
	mobs world.Mobs
}

func (g *Game) Setup() {
	world.EventBus.Subscribe((event.MobPosition{}).Type(), func(e event.Event) {
		evt := e.(*event.MobPosition)
		if mob := g.mobs.FindByID(evt.ID); mob != nil {
			mob.X = float64(evt.X)
			mob.Y = float64(evt.Y)
		}
	})
}

func (g *Game) Update() error {
	for _, mob := range g.mobs {
		mob.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, mob := range g.mobs {
		g.DrawMob(screen, mob)
	}
}

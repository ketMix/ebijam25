package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	mobs       []*Mob
	structures []*Structure
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.mobs[0].targetX = float64(x)
		g.mobs[0].targetY = float64(y)
	}
	if inpututil.IsKeyJustReleased(ebiten.Key1) {
		if len(g.mobs[0].individuals) > 10 {
			g.mobs[0].individuals = g.mobs[0].individuals[10:]
			g.structures = append(g.structures, &Structure{
				name: "village",
				x:    g.mobs[0].x,
				y:    g.mobs[0].y,
				rate: 240,
			})
		}
	} else if inpututil.IsKeyJustReleased(ebiten.Key2) {
		if len(g.mobs[0].individuals) > 10 {
			g.mobs[0].individuals = g.mobs[0].individuals[10:]
			g.mobs[0].AddStructure(&Structure{
				name: "mobile-village",
				x:    g.mobs[0].x,
				y:    g.mobs[0].y,
				rate: 480,
			})
		}
	}

	for _, mob := range g.mobs {
		mob.Update()
	}
	for _, structure := range g.structures {
		structure.Update()
	}

	eventBus.ProcessEvents()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, structure := range g.structures {
		structure.Draw(screen)
	}

	for _, mob := range g.mobs {
		mob.Draw(screen)
	}
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func main() {
	g := &Game{}
	ebiten.SetWindowSize(600, 600)
	ebiten.SetWindowTitle("Mobbox")

	g.mobs = append(g.mobs, &Mob{
		x: 300, y: 300,
		targetX: 300, targetY: 300,
	})

	for i := 0; i < 50; i++ {
		g.mobs[0].AddIndividual(&Individual{
			name: "chump",
		})
	}

	eventBus.Subscribe((&EventProduce{}).Type(), func(event Event) {
		e := event.(*EventProduce)
		// Find closest mob to x, y and add it to mob.
		closestMob := g.mobs[0]
		for _, mob := range g.mobs {
			if mob.x == e.individual.x && mob.y == e.individual.y {
				closestMob = mob
				break
			}
		}
		closestMob.AddIndividual(e.individual)
		fmt.Println("Produced individual:", e.individual.name, "at", e.individual.x, e.individual.y)
	})

	if err := loadImages(); err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

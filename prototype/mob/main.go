package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	mobs       []*Mob
	structures []*Structure
	resources  []*Resource
}

func (g *Game) Update() error {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.mobs[0].targetX = float64(x)
		g.mobs[0].targetY = float64(y)
	}
	if inpututil.IsKeyJustReleased(ebiten.Key1) {
		if len(g.mobs[0].Individuals()) > 10 {
			g.mobs[0].RemoveIndividuals(10)
			g.structures = append(g.structures, &Structure{
				name: "village",
				x:    g.mobs[0].x,
				y:    g.mobs[0].y,
				rate: 240,
			})
		}
	} else if inpututil.IsKeyJustReleased(ebiten.Key2) {
		if len(g.mobs[0].Individuals()) > 10 {
			g.mobs[0].RemoveIndividuals(10)
			g.mobs[0].AddParticipant(&Structure{
				name: "mobile-village",
				x:    g.mobs[0].x,
				y:    g.mobs[0].y,
				rate: 480,
			})
		}
	}

	for _, mob := range g.mobs {
		mob.Update(g)
	}
	for _, structure := range g.structures {
		structure.Update(nil)
	}

	eventBus.ProcessEvents()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, res := range g.resources {
		res.Draw(screen)
	}
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

	g.mobs = append(g.mobs, NewMob(300, 300))

	for i := 0; i < 50; i++ {
		g.mobs[0].AddParticipant(&Individual{
			name: "chump",
			x:    float64(300 + i%10*20),
			y:    float64(300 + i/10*10),
		})
	}

	g.resources = append(g.resources, NewResource(100, 100, 20))

	eventBus.Subscribe((&EventMerge{}).Type(), func(event Event) {
		e := event.(*EventMerge)
		a := g.FindMobById(e.from)
		b := g.FindMobById(e.to)
		if a == nil {
			fmt.Println("Merge event failed: from not found", e.from)
			return
		}
		if b == nil {
			fmt.Println("Merge event failed: to not found", e.to)
			return
		}
		if a.id == b.id {
			fmt.Println("Merge event ignored: same mob")
			return
		}
		b.participants = append(b.participants, a.participants...)
		fmt.Println(b.id, "now has", len(b.participants), "participants after merging with", a.id)
		g.RemoveMob(a)
	})

	eventBus.Subscribe((&EventProduce{}).Type(), func(event Event) {
		e := event.(*EventProduce)

		// This isn't correct, but whatever for now.
		if res := g.FindResourceWithin(e.structure.x, e.structure.y, 200); res == nil || res.food <= 0 {
			fmt.Println("Produce event failed: no resource found near", e.structure.x, e.structure.y)
			e.structure.failures++
			return
		} else {
			e.structure.failures = 0
			res.Deplete(1)
		}

		mob := NewMob(e.structure.x, e.structure.y)
		mob.AddParticipant(e.individual)

		// Find closest mob to x, y and set it as our target.
		/*closestMob := g.mobs[0]
		closestDist := float64(1<<63 - 1) // Start with a very large distance.
		for _, mob := range g.mobs {
			dx := mob.x - e.structure.x
			dy := mob.y - e.structure.y
			dist := dx*dx + dy*dy
			if dist < closestDist {
				closestDist = dist
				closestMob = mob
			}
		}
		mob.targetId = closestMob.id*/
		g.mobs = append(g.mobs, mob)
		fmt.Println("Produced individual:", e.individual.name, "at", e.individual.x, e.individual.y)
	})

	eventBus.Subscribe((&EventResourceDepleted{}).Type(), func(event Event) {
		e := event.(*EventResourceDepleted)

		res := g.FindResourceById(e.id)
		if res == nil {
			fmt.Println("ResourceDepleted event failed: resource not found", e.id)
			return
		}
		fmt.Println("Resource depleted:", res.id, "at", res.x, res.y)
		g.RemoveResource(res)
	})

	if err := loadImages(); err != nil {
		panic(err)
	}

	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}

func (g *Game) FindMobById(id int) *Mob {
	for _, mob := range g.mobs {
		if mob.id == id {
			return mob
		}
	}
	return nil
}

func (g *Game) RemoveMob(mob *Mob) {
	for i, m := range g.mobs {
		if m == mob {
			g.mobs = append(g.mobs[:i], g.mobs[i+1:]...)
			return
		}
	}
	fmt.Println("RemoveMob: Mob not found")
}

func (g *Game) FindResourceById(id int) *Resource {
	for _, res := range g.resources {
		if res.id == id {
			return res
		}
	}
	return nil
}

func (g *Game) RemoveResource(res *Resource) {
	for i, r := range g.resources {
		if r == res {
			g.resources = append(g.resources[:i], g.resources[i+1:]...)
			return
		}
	}
	fmt.Println("RemoveResource: Resource not found")
}

func (g *Game) FindResourceWithin(x, y float64, distance float64) *Resource {
	for _, res := range g.resources {
		dx := res.x - x
		dy := res.y - y
		if dx*dx+dy*dy <= distance*distance {
			return res
		}
	}
	return nil
}

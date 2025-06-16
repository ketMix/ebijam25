package main

import "math/rand"

type Director struct {
	tick     int
	nextTick int
}

func (d *Director) Update(g *Game) {
	d.tick++
	if d.tick >= d.nextTick {
		d.tick = 0
		d.nextTick = rand.Intn(600) + 600
		// Spawn some food.
		if len(g.resources) < 4 {
			eventBus.Publish(&EventResourceSpawn{
				x:    float64(rand.Intn(600)),
				y:    float64(rand.Intn(600)),
				food: rand.Intn(100) + 20,
			})
		}
	}
}

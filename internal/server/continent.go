package server

func (g *Game) UpdateContinent() {
	for _, mob := range g.Continent.Mobs {
		g.UpdateMob(mob)
	}

	/*for _, resource := range g.Resources {
		// Update resource logic
	}*/
}

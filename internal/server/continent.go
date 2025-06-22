package server

func (t *Table) UpdateContinent() {
	for _, mob := range t.Continent.Mobs {
		t.UpdateMob(mob)
	}

	/*for _, resource := range g.Resources {
		// Update resource logic
	}*/
}

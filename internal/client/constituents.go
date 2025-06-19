package client

import "github.com/ketMix/ebijam25/internal/world"

// All this will be removed if we switch to storing all schlub data in the ID.

type pendingMobConstituent struct {
	MobID       world.ID
	Constituent world.ID
}

type pendingConstituentsList []pendingMobConstituent

func (p *pendingConstituentsList) Add(mobID world.ID, constituent world.ID) {
	for _, pc := range *p {
		if pc.MobID == mobID && pc.Constituent == constituent {
			return
		}
	}
	// Add the new pending constituent.
	*p = append(*p, pendingMobConstituent{MobID: mobID, Constituent: constituent})
}

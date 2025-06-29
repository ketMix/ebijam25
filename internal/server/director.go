package server

import (
	"math/rand"

	"github.com/ketMix/ebijam25/internal/world"
)

const (
	MobTick           = 90
	ResourceTick      = 60
	MaxSchlubsToSpawn = 100
	MobStartingCount  = 200
)

type Timers struct {
	mobTimer      int
	resourceTimer int
}

type Director struct {
	table  *Table
	timers Timers
}

func NewDirector(t *Table) *Director {
	d := &Director{
		table: t,
		timers: Timers{
			mobTimer:      0,
			resourceTimer: 0,
		},
	}
	d.Setup()
	return d
}

func (d *Director) Setup() {
	for range MobStartingCount {
		d.AddMobs()
	}
}

func (d *Director) GetSpawnPosition() (float64, float64) {
	return rand.Float64() * world.ContinentPixelSpan, rand.Float64() * world.ContinentPixelSpan
}
func (d *Director) AddMobs() {
	// Spawn a family unit.
	t := d.table
	mobSchlubCount := d.table.Continent.Fate.NumGen.Intn(4) + 1
	posX, posY := d.GetSpawnPosition()

	fam := t.FamilyID.NextFamily()
	fam.SetKindID(int(world.SchlubKindVagrant)) // Set the kind to Vagrant for all random spawns.
	schlubs := fam.NextSchlubs(mobSchlubCount)
	// Actually randomize some of the schlubs to be monks or warriors.
	for i := range mobSchlubCount {
		if t.Continent.Fate.NumGen.Intn(100) < 20 {
			// 20% chance to make a schlub a monk or warrior
			if t.Continent.Fate.NumGen.Intn(100) < 50 {
				schlubs[i].SetKindID(int(world.SchlubKindMonk))
			} else {
				schlubs[i].SetKindID(int(world.SchlubKindWarrior))
			}
		}
	}
	t.FamilyID = schlubs[len(schlubs)-1]

	mob := t.Continent.NewMob(0, t.mobID.Next(), posX, posY)
	mob.OuterKind = world.SchlubKindVagrant // Set the outer kind to Vagrant for all random spawns.
	mob.AddSchlub(schlubs...)
	t.log.Debug("added mob", "id", mob.ID, "x", posX, "y", posY, "schlubs", len(mob.Schlubs))
}

func (d *Director) Update() {
	d.timers.mobTimer++
	d.timers.resourceTimer++

	if d.timers.mobTimer >= MobTick {
		d.AddMobs()
		d.timers.mobTimer = 0
	}

	if d.timers.resourceTimer >= ResourceTick {
		d.timers.resourceTimer = 0
	}
}

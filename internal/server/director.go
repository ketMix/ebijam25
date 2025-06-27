package server

import (
	"math/rand"

	"github.com/ketMix/ebijam25/internal/world"
)

const (
	MobTick           = 30
	ResourceTick      = 60
	MaxSchlubsToSpawn = 100
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
	// Create a fake mob a distance away to test mob visibility.
	t := d.table
	fam := t.FamilyID.NextFamily()
	t.FamilyID = fam.NextSchlub()
	mob := t.Continent.NewMob(2, t.mobID.Next(), 300, 300)
	mob.AddSchlub(fam)

	fam = t.FamilyID.NextFamily()
	schlubs := fam.NextSchlubs(50)
	t.FamilyID = schlubs[len(schlubs)-1]
	mob = t.Continent.NewMob(2, t.mobID.Next(), 200, 200)
	mob.AddSchlub(schlubs...)
}

func (d *Director) GetSpawnPosition() (float64, float64) {
	return rand.Float64() * world.ContinentPixelSpan, rand.Float64() * world.ContinentPixelSpan
}
func (d *Director) AddMobs() {
	t := d.table
	mobSchlubCount := int(rand.Float64()*MaxSchlubsToSpawn) + 1
	posX, posY := d.GetSpawnPosition()

	fam := t.FamilyID.NextFamily()
	schlubs := fam.NextSchlubs(mobSchlubCount)
	d.table.FamilyID = schlubs[len(schlubs)-1]

	mob := t.Continent.NewMob(2, t.mobID.Next(), posX, posY)
	mob.AddSchlub(schlubs...)
	t.log.Debug("added mob", "id", mob.ID, "position", posX, posY, "schlubs", len(mob.Schlubs))
}

func (d *Director) Update() {
	d.timers.mobTimer++
	d.timers.resourceTimer++

	if d.timers.mobTimer >= MobTick {
		// d.AddMobs()
		d.timers.mobTimer = 0
	}

	if d.timers.resourceTimer >= ResourceTick {
		d.timers.resourceTimer = 0
	}
}

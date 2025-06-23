package world

import (
	"math"
)

type Terrain int

const (
	TerrainNone Terrain = iota
	TerrainDirt
	TerrainGrass
	TerrainRockyDirt
	TerrainSand
	TerrainWater
	TerrainPines
	// TerrainMountain
	// TerrainForest
	// TerrainDesert
	// TerrainSwamp
	// TerrainSnow
	TerrainCount // Total number of terrain types
)

func NewTerrain(fate *Fate, x, y float64) Terrain {
	elevation := getElevation(fate, x, y)
	temperature := getTemperature(fate, x, y, elevation)
	moisture := getMoisture(fate, x, y)
	return getTerrain(elevation, temperature, moisture)
}

func (t Terrain) GetModifier() Modifier {
	switch t {
	case TerrainWater:
		return Modifier{
			Reason: "Your knees buckle against the dense current",
			Stats:  Stats{Agility: -5},
		}
	// case TerrainMountain:
	// 	return Modifier{
	// 		Reason: "Each new footing takes immense concentration",
	// 		Stats:  Stats{Agility: -5},
	// 	}
	case TerrainPines:
		return Modifier{
			Reason: "You thought you saw a little leprechaun man out of the corner of your eye...",
			Stats:  Stats{Luck: 1},
		}
	// case TerrainDesert:
	// 	return Modifier{
	// 		Reason: "The heat is unbearable, you feel like you're melting",
	// 		Stats:  Stats{Endurance: -2},
	// 	}
	// case TerrainSwamp:
	// 	return Modifier{
	// 		Reason: "The swamp is stinkin' you up.",
	// 		Stats:  Stats{Charisma: -2},
	// 	}
	// case TerrainSnow:
	// 	return Modifier{
	// 		Reason: "Your muscles seize up in the raw cold",
	// 		Stats:  Stats{Strength: -2},
	// 	}
	default:
		return Modifier{}
	}
}

func (t Terrain) ImageName() string {
	switch t {
	case TerrainNone:
		return "dirt"
	case TerrainDirt:
		return "dirt"
	case TerrainGrass:
		return "grass"
	case TerrainRockyDirt:
		return "rocky-dirt"
	case TerrainSand:
		return "sand"
	case TerrainWater:
		return "water"
	case TerrainPines:
		return "pines"
	default:
		return "dirt"
	}
}
func (t Terrain) String() string {
	switch t {
	case TerrainNone:
		return "None"
	case TerrainDirt:
		return "Dirt"
	case TerrainGrass:
		return "Grass"
	case TerrainRockyDirt:
		return "Rocky Dirt"
	case TerrainSand:
		return "Sand"
	case TerrainWater:
		return "Water"
	case TerrainPines:
		return "Pines"
	default:
		return "Unknown"
	}
}

func getElevation(fate *Fate, x, y float64) float64 {
	var value float64
	amplitude := 1.0
	frequency := 1.0
	maxValue := 0.0

	for range 6 {
		value += fate.Determine(
			x,
			y*frequency,
		) * amplitude

		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	value = (value/maxValue + 1) / 2
	value = math.Pow(value, 1.5)
	return value
}

func getTemperature(fate *Fate, x, y, elevation float64) float64 {
	temp := (fate.Determine(
		x,
		y,
	) + 1) / 2

	return temp * (1 - elevation)
}

func getMoisture(fate *Fate, x, y float64) float64 {
	moisture := (fate.Determine(
		x,
		y,
	) + 1) / 2
	return moisture
}

func getTerrain(elevation, temperature, moisture float64) Terrain {
	if elevation < 0.15 {
		return TerrainWater
	}
	if elevation < 0.2 {
		return TerrainSand
	}

	if elevation > 0.7 {
		return TerrainRockyDirt
	}

	if temperature > 0.6 {
		if moisture < 0.3 {
			return TerrainSand
		}
		if moisture < 0.5 {
			return TerrainRockyDirt
		}
		return TerrainSand
	}

	if moisture > 0.6 {
		return TerrainPines
	}
	if moisture > 0.4 {
		return TerrainGrass
	}
	return TerrainDirt
}

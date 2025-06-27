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
	TerrainCount // Total number of terrain types
)

func NewTerrain(fate *Fate, x, y float64) Terrain {
	elevation := getElevation(fate, x, y)
	temperature := getTemperature(fate, x, y, elevation)
	moisture := getMoisture(fate, x, y)
	return getTerrain(elevation, temperature, moisture)
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

	value += fate.Determine(
		x,
		y*frequency,
	) * amplitude

	maxValue += amplitude
	amplitude *= 0.5
	frequency *= 2.0

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
	if elevation < 0.05 {
		return TerrainWater
	}
	if elevation < 0.10 {
		return TerrainSand
	}

	if elevation > 0.9 {
		return TerrainRockyDirt
	}

	if temperature > 0.8 {
		if moisture < 0.3 {
			return TerrainSand
		}
		if moisture < 0.5 {
			return TerrainRockyDirt
		}
		return TerrainDirt
	}

	if moisture > 0.4 {
		return TerrainGrass
	}
	return TerrainDirt
}

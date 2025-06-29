package world

import (
	"math"
)

type Terrain int

const (
	TerrainNone Terrain = iota
	TerrainDirt
	TerrainGrass
	TerrainGrassyDirt
	TerrainGrassySand
	TerrainGrassyRocks
	TerrainRockyDirt
	TerrainRockySand
	TerrainSand
	TerrainRocks
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
	case TerrainGrassyDirt:
		return "grassy-dirt"
	case TerrainGrassyRocks:
		return "grassy-rocks"
	case TerrainGrassySand:
		return "grassy-sand"
	case TerrainRockyDirt:
		return "rocky-dirt"
	case TerrainRockySand:
		return "rocky-sand"
	case TerrainSand:
		return "sand"
	case TerrainRocks:
		return "rocks"
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
	case TerrainGrassyRocks:
		return "Grassy Rocks"
	case TerrainGrassySand:
		return "Grassy Sand"
	case TerrainGrassyDirt:
		return "Grassy Dirt"
	case TerrainRockyDirt:
		return "Rocky Dirt"
	case TerrainRockySand:
		return "Rocky Sand"
	case TerrainRocks:
		return "Rocks"
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
	if elevation < 0.3 {
		if moisture < 0.2 {
			return TerrainSand
		} else if moisture < 0.6 {
			return TerrainGrassySand
		} else if moisture <= 1.0 {
			return TerrainWater
		}
	} else if elevation < 0.6 {
		if moisture < 0.2 {
			return TerrainRocks
		} else if moisture < 0.3 {
			return TerrainRockyDirt
		} else if moisture < 0.4 {
			return TerrainDirt
		} else if moisture < 0.6 {
			return TerrainGrass
		} else if moisture <= 1.0 {
			return TerrainGrassyDirt
		}
	} else if elevation < 1.0 {
		if moisture < 0.2 {
			return TerrainGrass
		} else if moisture < 0.6 {
			return TerrainGrassyRocks
		} else if moisture < 0.8 {
			return TerrainRockyDirt
		} else if moisture <= 1.0 {
			return TerrainRocks
		}
	}
	return TerrainGrass
}

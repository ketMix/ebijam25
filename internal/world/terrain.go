package world

type Terrain int

const (
	TerrainNone Terrain = iota
	TerrainGrass
	TerrainWater
	TerrainMountain
	TerrainForest
	TerrainDesert
	TerrainSwamp
	TerrainSnow
)

func NewTerrain(sneed int) Terrain {
	// Apply a seed to the terrain generation logic if needed.
	return Terrain((sneed % 7) + 1) // Assuming 8 types of terrain
}

func (t Terrain) GetModifier() Modifier {
	switch t {
	case TerrainWater:
		return Modifier{
			Reason: "Your knees buckle against the dense current",
			Stats:  Stats{Agility: -5},
		}
	case TerrainMountain:
		return Modifier{
			Reason: "Each new footing takes immense concentration",
			Stats:  Stats{Agility: -5},
		}
	case TerrainForest:
		return Modifier{
			Reason: "You thought you saw a little leprechaun man out of the corner of your eye...",
			Stats:  Stats{Luck: 1},
		}
	case TerrainDesert:
		return Modifier{
			Reason: "The heat is unbearable, you feel like you're melting",
			Stats:  Stats{Endurance: -2},
		}
	case TerrainSwamp:
		return Modifier{
			Reason: "The swamp is stinkin' you up.",
			Stats:  Stats{Charisma: -2},
		}
	case TerrainSnow:
		return Modifier{
			Reason: "Your muscles seize up in the raw cold",
			Stats:  Stats{Strength: -2},
		}
	default:
		return Modifier{}
	}

}

func (t Terrain) String() string {
	switch t {
	case TerrainGrass:
		return "Grass"
	case TerrainWater:
		return "Water"
	case TerrainMountain:
		return "Mountain"
	case TerrainForest:
		return "Forest"
	case TerrainDesert:
		return "Desert"
	case TerrainSwamp:
		return "Swamp"
	case TerrainSnow:
		return "Snow"
	default:
		return "None"
	}
}

package progression

const (
	TechNameKnight = "knight"
	TechNameMonk   = "monk"
	TechNameNomad  = "nomad"
)

type TechSkill struct {
	name        string
	description string
	acquired    bool
	cost        int
	prereqs     []string
	usable      bool     // Whether the skill can be used directly by the player
	constructs  []string // Structures that can be constructed with this skill
}

type Tech struct {
	name        string
	description string
	skills      []TechSkill
}

type TechTree struct {
	techs map[string]Tech // Map of tech names to Tech objects
}

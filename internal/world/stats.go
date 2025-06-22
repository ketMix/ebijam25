package world

type Rank int

const (
	RankNone Rank = iota
	RankWooden
	RankBronze
	RankIron
	RankSteel
	RankGold
	RankPlatinum
	RankDiamond
)

func (r Rank) String() string {
	switch r {
	case RankNone:
		return "None"
	case RankWooden:
		return "Wooden"
	case RankBronze:
		return "Bronze"
	case RankIron:
		return "Iron"
	case RankSteel:
		return "Steel"
	case RankGold:
		return "Gold"
	case RankPlatinum:
		return "Platinum"
	case RankDiamond:
		return "Diamond"
	default:
		return "Unknown"
	}
}

func (r Rank) Clamp() Rank {
	if r < RankNone {
		return RankNone
	}
	if r > RankDiamond {
		return RankDiamond
	}
	return r
}

type Modifier struct {
	Stats
	Reason string // Reason for the modifier, e.g., "Potion", "Buff", etc.
}

type Stats struct {
	Strength  Rank // fightin
	Agility   Rank // running
	Charisma  Rank // persaudin
	Endurance Rank // survivin
	Luck      Rank // findin
}

func NewStats() *Stats {
	s := &Stats{}
	return s.clamp()
}

func (s *Stats) clamp() *Stats {
	s.Strength = Rank.Clamp(s.Strength)
	s.Agility = Rank.Clamp(s.Agility)
	s.Charisma = Rank.Clamp(s.Charisma)
	s.Endurance = Rank.Clamp(s.Endurance)
	s.Luck = Rank.Clamp(s.Luck)
	return s
}

func (s *Stats) Apply(other *Stats) *Stats {
	new := &Stats{
		Strength:  s.Strength + other.Strength,
		Agility:   s.Agility + other.Agility,
		Charisma:  s.Charisma + other.Charisma,
		Endurance: s.Endurance + other.Endurance,
		Luck:      s.Luck + other.Luck,
	}
	return new.clamp()

}

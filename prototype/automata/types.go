package main

// CellState represents the state of a cell in the automata grid
type CellState int

const (
	Empty CellState = iota // Empty cell
	Alive                  // Cell with a bullet
	Any                    // Wildcard - matches any state (for patterns)
)

// String returns string representation of CellState
func (cs CellState) String() string {
	switch cs {
	case Empty:
		return "Empty"
	case Alive:
		return "Alive"
	case Any:
		return "Any"
	default:
		return "Unknown"
	}
}

// Pattern3x3 represents a 3x3 pattern with condition and result
type Pattern3x3 struct {
	Condition [3][3]CellState `json:"condition"`
	Result    [3][3]CellState `json:"result"`
}

// NewPattern3x3 creates a new 3x3 pattern
func NewPattern3x3(condition, result [3][3]CellState) Pattern3x3 {
	return Pattern3x3{
		Condition: condition,
		Result:    result,
	}
}

// Matches checks if the pattern condition matches the given 3x3 area
func (p *Pattern3x3) Matches(area [3][3]CellState) bool {
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			conditionState := p.Condition[y][x]
			actualState := area[y][x]

			// Any matches everything
			if conditionState == Any {
				continue
			}

			// Exact match required for non-Any states
			if conditionState != actualState {
				return false
			}
		}
	}
	return true
}

// Apply applies the pattern transformation to the given area
func (p *Pattern3x3) Apply() [3][3]CellState {
	return p.Result
}

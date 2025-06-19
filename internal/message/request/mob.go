package request

import "github.com/ketMix/ebijam25/internal/message"

// Split represents a request to split a mob into a separate mob.
type Split struct {
	// FIXME: This will not be cloned properly in Decode.
	Schlubs []int `json:"schlubs"` // IDs of schlubs being split
}

// Type returns the type of the Split request.
func (s Split) Type() string {
	return "request-split"
}

// Move represents a request to move a mob towards a new position.
type Move struct {
	X int `json:"x"` // X coordinate to move to
	Y int `json:"y"` // Y coordinate to move to
}

// Type returns the type of the Move request.
func (m Move) Type() string {
	return "request-move"
}

// Formation represents a request to adjust the formation of a mob to have the schlubs organized from center outwards.
type Formation struct {
	// FIXME: This will not be cloned properly in Decode.
	Order []string `json:"order"` // Order of schlubs from center outwards.
}

// Type returns the type of the Formation request.
func (f Formation) Type() string {
	return "request-formation"
}

// Construct represents a request to construct a specific type of structure.
type Construct struct {
	Structure string `json:"structure"` // Type of structure to construct (e.g., "monastery", "barracks")
}

// Type returns the type of the Construct request.
func (c Construct) Type() string {
	return "request-construct"
}

func init() {
	message.Register(&Split{})
	message.Register(&Move{})
	message.Register(&Formation{})
	message.Register(&Construct{})
}

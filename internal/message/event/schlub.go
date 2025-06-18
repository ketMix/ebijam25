package event

import "github.com/ketMix/ebijam25/internal/message"

// SchlubCreate represents an event where a new schlub is created with a unique ID and type.
type SchlubCreate struct {
	ID   int    `json:"id"`   // ID of the schlub
	Kind string `json:"kind"` // Type of schlub (e.g., "warrior", "monk")
}

// Type returns the type of the SchlubCreate event.
func (s SchlubCreate) Type() string {
	return "schlub-create"
}

// SchlubCreateMany represents an event where multiple schlubs of the same type are created at once.
type SchlubCreateMany struct {
	Kind string `json:"kind"` // Type of schlub (e.g., "warrior", "monk")
	IDs  []int  `json:"ids"`  // IDs of the schlubs being created
}

// Type returns the type of the SchlubCreateMany event.
func (s SchlubCreateMany) Type() string {
	return "schlub-createMany"
}

// SchlubCreateList represents an event where multiple schlubs are created at once.
type SchlubCreateList struct {
	Schlubs []SchlubCreate `json:"schlubs"` // List of schlubs being created
}

// Type returns the type of the SchlubCreateList event.
func (s SchlubCreateList) Type() string {
	return "schlub-createList"
}

// SchlubPlace represents an event where a schlub is placed at a specific location on the map.
type SchlubPlace struct {
	ID int `json:"id"` // ID of the schlub being placed
	X  int `json:"x"`  // X coordinate of the schlub
	Y  int `json:"y"`  // Y coordinate of the schlub
}

// Type returns the type of the SchlubPlace event.
func (s SchlubPlace) Type() string {
	return "schlub-place"
}

func init() {
	message.Register(SchlubCreate{})
	message.Register(SchlubCreateMany{})
	message.Register(SchlubCreateList{})
	message.Register(SchlubPlace{})
}

package event

// SchlubCreate represents an event where a new schlub is created with a unique ID and type.
type SchlubCreate struct {
	ID   int    `json:"id"`   // ID of the schlub
	Kind string `json:"kind"` // Type of schlub (e.g., "warrior", "monk")
}

// Type returns the type of the SchlubCreate event.
func (s SchlubCreate) Type() string {
	return "schlub-create"
}

// SchlubCreateMany represents an event where multiple schlubs of the same type are created at once. It is presumed that ID is from ID to ID+Count-1.
type SchlubCreateMany struct {
	SchlubCreate
	Count int `json:"count"` // Number of schlubs of this type
}

// Type returns the type of the SchlubCreateMany event.
func (s SchlubCreateMany) Type() string {
	return "schlub-createMany"
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
	Register(SchlubCreate{})
	Register(SchlubCreateMany{})
	Register(SchlubPlace{})
}

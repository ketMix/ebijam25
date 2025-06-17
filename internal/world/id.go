package world

// IDGenerator is a simple ID generator that provides unique IDs.
type IDGenerator struct {
	currentID int
}

// NewIDGenerator creates a new ID generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{currentID: 0}
}

// Next generates the next unique ID.
func (gen *IDGenerator) Next() int {
	gen.currentID++
	return gen.currentID
}

// Reset resets the ID generator to start from 0 again.
func (gen *IDGenerator) Reset() {
	gen.currentID = 0
}

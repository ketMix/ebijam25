package world

import (
	"fmt"
)

// ID represents a unique identifier.
type ID = int

// IDGenerator is a simple ID generator that provides unique IDs.
type IDGenerator struct {
	currentID ID
}

// NewIDGenerator creates a new ID generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{currentID: 0}
}

// Next generates the next unique ID.
func (gen *IDGenerator) Next() ID {
	gen.currentID++
	return gen.currentID
}

// Reset resets the ID generator to start from 0 again.
func (gen *IDGenerator) Reset() {
	gen.currentID = 0
}

/*
9-bit family ID (511)
10-bit constituents ID (1023)
3-bit mob ID (7)
5-bit item ID (31)
5-bit level (31)
*/

type SchlubID int

func (s SchlubID) String() string {
	return fmt.Sprintf("family: %d, schlub %d, kind: %d, item: %d, bits: %s",
		s.FamilyID(),
		s.SchlubID(),
		s.KindID(),
		s.ItemID(),
		s.BitsAsString(),
	)
}

// NextFamily gets the next family ID from the current schlub ID. It returns a new SchlubID with the family incremented from the former and all other fields zeroed out.
func (s SchlubID) NextFamily() SchlubID {
	familyID := int(s) >> 22
	familyID++
	if familyID > 511 {
		familyID = 0
	}
	return SchlubID(familyID << 22)
}

func (s SchlubID) FamilyID() int {
	// Extract the 9-bit family ID from the SchlubID
	return (int(s) >> 22) & 0x1FF
}

func (s SchlubID) NextSchlub() SchlubID {
	// Keep the 9-bit family id and increment the 10-bit schlub id
	schlubID := (int(s) >> 12) & 0x3FF
	schlubID++
	if schlubID > 1023 {
		schlubID = 0
	}
	return SchlubID((int(s) & 0xFFC00000) | (schlubID << 12))
}

func (s SchlubID) SchlubID() int {
	// Extract the 10-bit schlub ID from the SchlubID
	return (int(s) >> 12) & 0x3FF
}

func (s SchlubID) BitsAsString() string {
	// Convert the SchlubID to a 32-bit binary string representation
	return fmt.Sprintf("%032b", int(s))
}

func (s SchlubID) KindID() int {
	return (int(s) >> 9) & 0x7
}
func (s *SchlubID) SetKindID(kind int) {
	// Set the 3-bit kind ID in the SchlubID
	if kind < 0 || kind > 7 {
		panic("kind must be between 0 and 7")
	}
	*s = SchlubID((int(*s) & 0xFFFFF1FF) | (kind << 9))
}

func (s SchlubID) ItemID() int {
	// Extract the 5-bit item ID from the SchlubID
	return (int(s) >> 4) & 0x1F
}

func (s *SchlubID) SetItemID(item int) {
	// Set the 5-bit item ID in the SchlubID
	if item < 0 || item > 31 {
		panic("item must be between 0 and 31")
	}
	*s = SchlubID((int(*s) & 0xFFFFFE1F) | (item << 4))
}

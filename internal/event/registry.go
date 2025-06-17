package event

import (
	"encoding/json"
)

// EventI is an interface that all event types must implement.
type EventI interface {
	Type() string
}

// TypedEvent is a wrapper for events that includes the event type and the event data.
type TypedEvent struct {
	Type string          `json:"type"` // Type of the event
	Data json.RawMessage `json:"data"` // Data of the event
}

var registry = map[string]EventI{}

// Register adds an EventI instance to the registry. It panics if the event type is already registered.
func Register(event EventI) {
	if _, exists := registry[event.Type()]; exists {
		panic("event type already registered: " + event.Type())
	}
	registry[event.Type()] = event
}

// Encode takes an EventI instance and returns a byte slice containing the JSON-encoded event data.
func Encode(event EventI) ([]byte, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Add the event type to the JSON data
	typedEvent := TypedEvent{
		Type: event.Type(),
		Data: data,
	}

	data, err = json.Marshal(typedEvent)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Decode takes a byte slice containing a JSON-encoded event and returns the corresponding EventI instance.
func Decode(data []byte) (EventI, error) {
	var typedEvent TypedEvent
	if err := json.Unmarshal(data, &typedEvent); err != nil {
		return nil, err
	}

	event, exists := registry[typedEvent.Type]
	if !exists {
		return nil, nil // or return an error if you prefer
	}

	event2 := event // Create a new instance of the event type. FIXME: This won't clone contained slices...

	// Unmarshal the data into the specific event type
	err := json.Unmarshal(typedEvent.Data, event2)
	if err != nil {
		return nil, err
	}

	return event, nil
}

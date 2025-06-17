package message

import (
	"encoding/json"
)

// MessageI is an interface that all message types must implement.
type MessageI interface {
	Type() string
}

// TypedMessage is a wrapper for messages that includes the message type and the message data.
type TypedMessage struct {
	Type string          `json:"type"` // Type of the message
	Data json.RawMessage `json:"data"` // Data of the message
}

var registry = map[string]MessageI{}

// Register adds an MessageI instance to the registry. It panics if the message type is already registered.
func Register(message MessageI) {
	if _, exists := registry[message.Type()]; exists {
		panic("message type already registered: " + message.Type())
	}
	registry[message.Type()] = message
}

// Encode takes an MessageI instance and returns a byte slice containing the JSON-encoded message data.
func Encode(message MessageI) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	// Add the message type to the JSON data
	typedMessage := TypedMessage{
		Type: message.Type(),
		Data: data,
	}

	data, err = json.Marshal(typedMessage)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Decode takes a byte slice containing a JSON-encoded message and returns the corresponding MessageI instance.
func Decode(data []byte) (MessageI, error) {
	var typedMessage TypedMessage
	if err := json.Unmarshal(data, &typedMessage); err != nil {
		return nil, err
	}

	message, exists := registry[typedMessage.Type]
	if !exists {
		return nil, nil // or return an error if you prefer
	}

	message2 := message // Create a new instance of the message type. FIXME: This won't clone contained slices...

	// Unmarshal the data into the specific message type
	err := json.Unmarshal(typedMessage.Data, message2)
	if err != nil {
		return nil, err
	}

	return message, nil
}

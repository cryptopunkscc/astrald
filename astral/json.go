package astral

import "encoding/json"

// JSONDecodeAdapter is used to decode astral objects from a JSON. Use Blueprints.RefineJSON to convert this to Object.
type JSONDecodeAdapter struct {
	Type    string
	Object  json.RawMessage `json:",omitempty"`
	Payload []byte          `json:",omitempty"`
}

// JSONEncodeAdapter is used to encode astral objects to JSON.
type JSONEncodeAdapter struct {
	Type    string
	Object  any    `json:",omitempty"`
	Payload []byte `json:",omitempty"`
}

var jsonNull = []byte("null")

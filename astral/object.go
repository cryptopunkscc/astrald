package astral

import (
	"encoding/json"
	"io"
)

// Object defines the basic interface of an astral object. An object must have a unique type and must be able to
// write/read its payload (the type is outside the payload) to/from a stream.
type Object interface {
	ObjectType() string
	io.WriterTo
	io.ReaderFrom
}

type JSONObject interface {
	Object
	json.Marshaler
	json.Unmarshaler
}

// JSONAdapter is used as a generic container for JSON-encoded Objects.
type JSONAdapter struct {
	Type   string
	Object json.RawMessage `json:",omitempty"`
}

var jsonNull = []byte("null")

// ResolveObjectID calculates the id of the object
func ResolveObjectID(obj Object) (objectID *ObjectID, err error) {
	w := NewWriteResolver(nil)

	// write the astral stamp
	_, err = Stamp{}.WriteTo(w)
	if err != nil {
		return
	}

	// write the object type
	_, err = ObjectType(obj.ObjectType()).WriteTo(w)
	if err != nil {
		return
	}

	// write the object payload
	_, err = obj.WriteTo(w)
	if err != nil {
		return
	}

	return w.Resolve(), nil
}

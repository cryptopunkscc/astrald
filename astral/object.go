package astral

import (
	"bytes"
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

// Write writes the object to the writer using DefaultBlueprints
func Write(w io.Writer, obj Object) (_ int64, err error) {
	return DefaultBlueprints.Write(w, obj)
}

// WriteJSON writes the object in its JSON form to the writer
func WriteJSON(w io.Writer, obj Object) (err error) {
	enc := json.NewEncoder(w)
	switch obj := obj.(type) {
	case *RawObject:
		err = enc.Encode(&JSONEncodeAdapter{
			Type:    obj.ObjectType(),
			Payload: obj.Payload,
		})

	default:
		err = enc.Encode(&JSONEncodeAdapter{
			Type:   obj.ObjectType(),
			Object: obj,
		})
	}

	return
}

// Pack writes the object in its short form to a buffer and returns the buffer
func Pack(obj Object) ([]byte, error) {
	return DefaultBlueprints.Pack(obj)
}

// PackJSON writes the object in its JSON form to a buffer and returns the buffer
func PackJSON(obj Object) (_ []byte, err error) {
	var buf = &bytes.Buffer{}
	err = WriteJSON(buf, obj)
	return buf.Bytes(), err
}

// ObjectReader is obsolete.
type ObjectReader interface {
	ReadObject() (Object, int64, error)
}

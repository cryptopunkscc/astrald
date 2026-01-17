package astral

import (
	"encoding"
	"encoding/json"
	"io"
)

// Object defines the basic interface of an astral object. An object must have a unique type and must be able to
// write/read its payload (the type is outside the payload) to/from a stream.
type Object interface {
	ObjectType() string
	WriteTo(io.Writer) (n int64, err error)  // io.WriterTo
	ReadFrom(io.Reader) (n int64, err error) // io.ReaderFrom

}

// JSONObject is an Object that supports JSON encoding and decoding.
type JSONObject interface {
	Object
	json.Marshaler
	json.Unmarshaler
}

// TextObject is an Object that supports text encoding and decoding.
type TextObject interface {
	Object
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

// JSONAdapter is used as a generic container for JSON-encoded Objects.
type JSONAdapter struct {
	Type   string
	Object json.RawMessage `json:",omitempty"`
}

var jsonNull = []byte("null")

type endecConfig struct {
	Blueprints *Blueprints
	Encoder    TypeEncoder
	Decoder    TypeDecoder
}

type ConfigFunc func(*endecConfig)

// Encode encodes the object to the writer
func Encode(w io.Writer, obj Object, config ...ConfigFunc) (n int64, err error) {
	cfg := makeConfig(config...)

	n, err = cfg.Encoder(w, obj.ObjectType())
	if err != nil {
		return
	}

	m, err := obj.WriteTo(w)
	n += m

	return
}

// Decode decodes an object from the reader
func Decode(r io.Reader, config ...ConfigFunc) (object Object, n int64, err error) {
	cfg := makeConfig(config...)

	typ, n, err := cfg.Decoder(r)
	if err != nil {
		return
	}

	object = cfg.Blueprints.New(typ)
	if object == nil {
		return nil, n, newErrBlueprintNotFound(typ)
	}

	m, err := object.ReadFrom(r)
	n += m

	if err != nil {
		object = nil
	}

	return
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

func WithEncoder(enc TypeEncoder) ConfigFunc {
	return func(cfg *endecConfig) {
		cfg.Encoder = enc
	}
}

func WithDecoder(dec TypeDecoder) ConfigFunc {
	return func(cfg *endecConfig) {
		cfg.Decoder = dec
	}
}

func WithBlueprints(bp *Blueprints) ConfigFunc {
	return func(cfg *endecConfig) {
		cfg.Blueprints = bp
	}
}

func makeConfig(config ...ConfigFunc) *endecConfig {
	cfg := &endecConfig{
		Encoder:    ShortTypeEncoder,
		Decoder:    ShortTypeDecoder,
		Blueprints: DefaultBlueprints(),
	}
	for _, f := range config {
		f(cfg)
	}
	return cfg
}

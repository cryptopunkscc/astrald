package astral

import (
	"encoding/json"
	"io"
)

var _ Object = &RawObject{}

// RawObject is an Object that holds an unparsed payload. See Blueprints.Refine on how to parse these objects.
type RawObject struct {
	Type    string
	Payload []byte
}

func (raw *RawObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(raw.Payload)
}

func (raw *RawObject) ObjectType() string {
	return raw.Type
}

func (raw *RawObject) WriteTo(w io.Writer) (n int64, err error) {
	m, err := w.Write(raw.Payload)
	return int64(m), err
}

func (raw *RawObject) ReadFrom(r io.Reader) (n int64, err error) {
	raw.Payload, err = io.ReadAll(r)
	n = int64(len(raw.Payload))
	return
}

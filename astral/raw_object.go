package astral

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
)

// RawObject is an Object that holds an unparsed payload. See Blueprints.Refine on how to parse these objects.
type RawObject struct {
	Type    string
	Payload []byte
}

var _ Object = &RawObject{}

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

func (raw *RawObject) MarshalText() (text []byte, err error) {
	if len(raw.Payload) == 0 {
		return []byte{}, nil
	}

	s := base64.StdEncoding.EncodeToString(raw.Payload)

	return []byte(s), nil
}

func (raw *RawObject) UnmarshalText(text []byte) (err error) {
	if len(bytes.TrimSpace(text)) == 0 {
		return nil
	}

	raw.Payload, err = base64.StdEncoding.DecodeString(string(text))

	return
}

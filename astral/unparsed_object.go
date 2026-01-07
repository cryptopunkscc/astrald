package astral

import (
	"errors"
	"fmt"
	"io"
)

// UnparsedObject is a struct that holds an unparsed object. Use Blueprints.Parse to parse.
type UnparsedObject struct {
	Type    string
	Payload []byte
}

var _ Object = &UnparsedObject{}

func NewUnparsedObject(typ string, payload []byte) *UnparsedObject {
	return &UnparsedObject{Type: typ, Payload: payload}
}

func (unparsed *UnparsedObject) ObjectType() string {
	return unparsed.Type
}

// binary

func (unparsed *UnparsedObject) WriteTo(w io.Writer) (n int64, err error) {
	m, err := w.Write(unparsed.Payload)
	return int64(m), err
}

func (unparsed *UnparsedObject) ReadFrom(r io.Reader) (n int64, err error) {
	unparsed.Payload, err = io.ReadAll(r)
	n = int64(len(unparsed.Payload))
	return
}

// json

func (unparsed UnparsedObject) MarshalJSON() ([]byte, error) {
	return nil, errors.New("cannot marshal an unparsed object")
}

func (unparsed *UnparsedObject) UnmarshalJSON(bytes []byte) error {
	return errors.New("cannot unmarshal an unparsed object")
}

// NOTE: text interface has to be omitted to force base64 encoding in channels

func (unparsed UnparsedObject) String() string {
	return fmt.Sprintf("[unparsed %s]", unparsed.Type)
}

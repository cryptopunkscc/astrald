package astral

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
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

func (raw *RawObject) MarshalText() (text []byte, err error) {
	var buf = &bytes.Buffer{}

	var p = ""
	if len(raw.Payload) > 0 {
		p = base64.StdEncoding.EncodeToString(raw.Payload)
	}

	fmt.Fprintf(buf, "{{%s:%s}}", raw.Type, p)

	return buf.Bytes(), nil
}

func (raw *RawObject) UnmarshalText(text []byte) (err error) {
	var s = string(text)
	s, ok := strings.CutPrefix(s, "{{")
	if !ok {
		return errors.New("invalid format")
	}
	s, ok = strings.CutSuffix(s, "}}")
	if !ok {
		return errors.New("invalid format")
	}
	p := strings.SplitN(s, ":", 2)
	if len(p) != 2 {
		return errors.New("invalid format")
	}

	raw.Type = p[0]
	raw.Payload, err = base64.StdEncoding.DecodeString(p[1])
	return
}

// ToRaw converts an Object to a RawObject
func ToRaw(obj Object) (raw *RawObject, err error) {
	var buf = &bytes.Buffer{}
	_, err = obj.WriteTo(buf)
	raw = &RawObject{
		Type:    obj.ObjectType(),
		Payload: buf.Bytes(),
	}
	return
}

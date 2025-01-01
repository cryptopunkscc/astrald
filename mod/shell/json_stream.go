package shell

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type JSONStream struct {
	rw  io.ReadWriter
	enc *json.Encoder
	dec *json.Decoder
}

func NewJSONStream(rw io.ReadWriter) *JSONStream {
	return &JSONStream{
		rw:  rw,
		enc: json.NewEncoder(rw),
		dec: json.NewDecoder(rw),
	}
}

func (stream JSONStream) ReadObject() (astral.Object, int64, error) {
	var j jsonWrapper

	err := stream.dec.Decode(&j)
	if err != nil {
		return nil, 0, err
	}

	bp := astral.ExtractBlueprints(stream.rw)

	var o any

	switch {
	case len(j.Data) > 0:
		o = bp.Make(j.Type)
		if o == nil {
			return nil, 0, fmt.Errorf("unable to parse object type: %s", j.Type)
		}
		err = json.NewDecoder(bytes.NewReader(j.Data)).Decode(&o)

	case len(j.Payload) > 0:
		o = &astral.RawObject{Type: j.Type}
		err = json.NewDecoder(bytes.NewReader(j.Data)).Decode(&o)
	}
	if err != nil {
		return nil, 0, err
	}

	object := o.(astral.Object)

	return object, 0, err
}

func (stream JSONStream) WriteObject(object astral.Object) (n int64, err error) {
	var data []byte
	var wrapped = JSONWrapper{object}

	data, err = wrapped.MarshalJSON()
	if err != nil {
		return
	}
	data = append(data, '\n')

	m, err := stream.rw.Write(data)

	return int64(m), err
}

type JSONWrapper struct {
	astral.Object
}

type jsonWrapper struct {
	Type    string
	Data    json.RawMessage `json:"Data,omitempty"`
	Payload []byte          `json:"Payload,omitempty"`
}

func (w JSONWrapper) MarshalJSON() (_ []byte, err error) {
	var p []byte
	p, err = json.Marshal(w.Object)
	if err != nil {
		return
	}

	var j = jsonWrapper{
		Type: w.Object.ObjectType(),
		Data: json.RawMessage(p),
	}

	return json.Marshal(j)
}

package channel

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type JSONWriter struct {
	w   io.Writer
	enc *json.Encoder
}

var _ Writer = &JSONWriter{}

func NewJSONWriter(w io.Writer) *JSONWriter {
	return &JSONWriter{w: w, enc: json.NewEncoder(w)}
}

func (w JSONWriter) Write(object astral.Object) (err error) {
	switch obj := object.(type) {
	case *astral.RawObject:
		err = w.enc.Encode(&astral.JSONEncodeAdapter{
			Type:    obj.ObjectType(),
			Payload: obj.Payload,
		})

	default:
		err = w.enc.Encode(&astral.JSONEncodeAdapter{
			Type:   obj.ObjectType(),
			Object: obj,
		})
	}

	return
}

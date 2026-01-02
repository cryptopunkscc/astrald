package channel

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// JSONSender writes a stream of astral.Objects encoded as JSON lines to the underlying io.Writer.
type JSONSender struct {
	w   io.Writer
	enc *json.Encoder
}

var _ Sender = &JSONSender{}

func NewJSONSender(w io.Writer) *JSONSender {
	return &JSONSender{w: w, enc: json.NewEncoder(w)}
}

func (w JSONSender) Send(object astral.Object) (err error) {
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

package channel

import (
	"encoding/json"
	"errors"
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
	switch object.(type) {
	case *astral.UnparsedObject:
		return errors.New("cannot send unparsed objects over JSON")
	}

	j := astral.JSONAdapter{
		Type: object.ObjectType(),
	}

	j.Object, err = json.Marshal(object)
	if err != nil {
		return err
	}

	return w.enc.Encode(&j)
}

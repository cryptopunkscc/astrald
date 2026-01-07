package channel

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// JSONReceiver reads a stream of astral.Objects encoded as JSON lines from the underlying io.Reader.
type JSONReceiver struct {
	r   io.Reader
	bp  *astral.Blueprints
	dec *json.Decoder
}

var _ Receiver = &JSONReceiver{}

func NewJSONReceiver(r io.Reader) *JSONReceiver {
	return &JSONReceiver{
		r:   r,
		bp:  astral.ExtractBlueprints(r),
		dec: json.NewDecoder(r),
	}
}

func (r JSONReceiver) Receive() (object astral.Object, err error) {
	var jsonObj astral.JSONDecodeAdapter

	err = r.dec.Decode(&jsonObj)
	if err != nil {
		return nil, err
	}

	object = r.bp.Make(jsonObj.Type)
	if object == nil {
		return nil, astral.ErrBlueprintNotFound{Type: jsonObj.Type}
	}

	if jsonObj.Object != nil {
		err = json.Unmarshal(jsonObj.Object, &object)
	}
	if err != nil {
		return nil, err
	}

	return
}

package channel

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// JSONReceiver reads a stream of astral.Objects encoded as JSON lines from the underlying io.Reader.
type JSONReceiver struct {
	r         io.Reader
	dec       *json.Decoder
	streamErr error
}

var _ Receiver = &JSONReceiver{}

func NewJSONReceiver(r io.Reader) *JSONReceiver {
	return &JSONReceiver{
		r:   r,
		dec: json.NewDecoder(r),
	}
}

// AllowUnparsed is not honored on JSON streams. Although json.Decoder is document-framed,
// the policy here is fail-fast: the first non-nil error latches and subsequent Receive()
// calls return it without touching the reader.
func (r *JSONReceiver) Receive() (object astral.Object, err error) {
	if r.streamErr != nil {
		return nil, r.streamErr
	}
	defer func() {
		if err != nil {
			r.streamErr = err
		}
	}()

	var jsonObj astral.JSONAdapter

	err = r.dec.Decode(&jsonObj)
	if err != nil {
		return nil, err
	}

	object = astral.New(jsonObj.Type)
	if object == nil {
		return nil, fmt.Errorf("%w: %w: %s", astral.ErrStreamCorrupted, astral.ErrBlueprintNotFound, jsonObj.Type)
	}

	if jsonObj.Object != nil {
		err = json.Unmarshal(jsonObj.Object, &object)
	}
	if err != nil {
		return nil, err
	}

	return
}

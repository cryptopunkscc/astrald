package channel

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// JSONReader reads a stream of astral.Objects encoded as JSON lines from the underlying io.Reader.
type JSONReader struct {
	r   io.Reader
	bp  *astral.Blueprints
	dec *json.Decoder
}

var _ Reader = &JSONReader{}

func NewJSONReader(r io.Reader) *JSONReader {
	return &JSONReader{
		r:   r,
		bp:  astral.ExtractBlueprints(r),
		dec: json.NewDecoder(r),
	}
}

func (r JSONReader) Read() (object astral.Object, err error) {
	var jsonObj astral.JSONDecodeAdapter

	err = r.dec.Decode(&jsonObj)
	if err != nil {
		return
	}

	return r.bp.RefineJSON(&jsonObj)
}

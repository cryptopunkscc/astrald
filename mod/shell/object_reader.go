package shell

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.ObjectReader = &ObjectReader{}

// ObjectReader reads canonical, 32-bit packed objects
type ObjectReader struct {
	r io.Reader
}

func NewObjectReader(r io.Reader) *ObjectReader {
	return &ObjectReader{
		r: r,
	}
}

func (r ObjectReader) ReadObject() (object astral.Object, n int64, err error) {
	var b astral.Bytes32

	n, err = b.ReadFrom(r.r)
	if err != nil {
		return
	}

	object, _, err = astral.ExtractBlueprints(r.r).Read(bytes.NewReader(b))

	return
}

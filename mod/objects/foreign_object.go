package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &ForeignObject{}

// ForeignObject is an astral.Object that holds objects of unknown structure.
type ForeignObject struct {
	Type    string
	Payload []byte
}

func (fo *ForeignObject) ObjectType() string {
	return fo.Type
}

func (fo *ForeignObject) WriteTo(w io.Writer) (n int64, err error) {
	m, err := w.Write(fo.Payload)
	return int64(m), err
}

func (fo *ForeignObject) ReadFrom(r io.Reader) (n int64, err error) {
	fo.Payload, err = io.ReadAll(r)
	n = int64(len(fo.Payload))
	return
}

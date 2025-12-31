package channel

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type BinaryWriter struct {
	w io.Writer
}

var _ Writer = &BinaryWriter{}

func NewBinaryWriter(w io.Writer) *BinaryWriter {
	return &BinaryWriter{w: w}
}

func (w BinaryWriter) Write(object astral.Object) (err error) {
	var frame []byte

	frame, err = astral.Pack(object)
	if err != nil {
		return
	}
	_, err = astral.Bytes16(frame).WriteTo(w.w)
	return
}

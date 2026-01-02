package channel

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// BinarySender writes a stream of astral.Objects to the underlying io.Writer.
type BinarySender struct {
	w io.Writer
}

var _ Sender = &BinarySender{}

func NewBinarySender(w io.Writer) *BinarySender {
	return &BinarySender{w: w}
}

func (w BinarySender) Send(object astral.Object) (err error) {
	var frame []byte

	frame, err = astral.Pack(object)
	if err != nil {
		return
	}
	_, err = astral.Bytes16(frame).WriteTo(w.w)
	return
}

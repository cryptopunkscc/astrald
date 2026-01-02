package channel

import (
	"bytes"
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
	// write the object type
	_, err = astral.String8(object.ObjectType()).WriteTo(w.w)
	if err != nil {
		return
	}

	// buffer the payload
	var buf = bytes.NewBuffer(nil)
	_, err = object.WriteTo(buf)
	if err != nil {
		return
	}

	// write the buffer with 32-bit length prefix
	_, err = astral.Bytes32(buf.Bytes()).WriteTo(w.w)

	return
}

package channel

import (
	"bytes"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// BinaryReceiver reads a stream of astral.Objects from the underlying io.Reader.
type BinaryReceiver struct {
	bp *astral.Blueprints
	r  io.Reader
}

var _ Receiver = &BinaryReceiver{}

func NewBinaryReceiver(r io.Reader) *BinaryReceiver {
	return &BinaryReceiver{r: r, bp: astral.ExtractBlueprints(r)}
}

func (b BinaryReceiver) Receive() (object astral.Object, err error) {
	var frame astral.Bytes16

	_, err = frame.ReadFrom(b.r)
	if err != nil {
		return
	}

	object, _, err = b.bp.Read(bytes.NewReader(frame))

	return
}

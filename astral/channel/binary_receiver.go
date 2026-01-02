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
	// read the object type
	var objectType astral.ObjectType
	_, err = objectType.ReadFrom(b.r)
	if err != nil {
		return
	}

	if len(objectType) == 0 {
		object = &astral.Blob{}
	} else {
		object = b.bp.Make(objectType.String())
		if object == nil {
			return nil, ErrUnknownObject
		}
	}

	// read the object payload
	var buf astral.Bytes32
	_, err = buf.ReadFrom(b.r)
	if err != nil {
		return
	}

	// decode the payload
	_, err = object.ReadFrom(bytes.NewReader(buf))

	return
}

package channel

import (
	"bytes"
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// BinaryReceiver reads a stream of astral.Objects from the underlying io.Reader.
type BinaryReceiver struct {
	bp            *astral.Blueprints
	r             io.Reader
	AllowUnparsed bool
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

	// read the object payload
	var buf astral.Bytes32
	_, err = buf.ReadFrom(b.r)
	if err != nil {
		return nil, err
	}

	if len(objectType) == 0 {
		object = &astral.Blob{}
	} else {
		object = b.bp.Make(objectType.String())
	}

	if object == nil {
		if b.AllowUnparsed {
			return &astral.RawObject{Type: objectType.String(), Payload: buf}, nil
		}

		return nil, astral.ErrBlueprintNotFound{Type: objectType.String()}
	}

	// decode the payload
	_, err = object.ReadFrom(bytes.NewReader(buf))
	switch {
	case err == nil:
		return

	case errors.Is(err, astral.ErrBlueprintNotFound{}) && b.AllowUnparsed:
		return &astral.RawObject{Type: objectType.String(), Payload: buf}, nil

	default:
		return nil, err
	}

}

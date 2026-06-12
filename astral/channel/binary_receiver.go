package channel

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// BinaryReceiver reads a stream of astral.Objects from the underlying io.Reader.
type BinaryReceiver struct {
	r             io.Reader
	AllowUnparsed bool
}

var _ Receiver = &BinaryReceiver{}

func NewBinaryReceiver(r io.Reader) *BinaryReceiver {
	return &BinaryReceiver{r: r}
}

// Receive reads one type-prefixed, length-framed object.
// An empty type yields a *astral.Blob; an unknown type yields an UnparsedObject only if AllowUnparsed.
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

	// if there's no type, it's a blob
	if len(objectType) == 0 {
		return (*astral.Blob)(&buf), nil
	}

	object = astral.New(objectType.String())
	if object == nil {
		if b.AllowUnparsed {
			return astral.NewUnparsedObject(objectType.String(), buf), nil
		}

		return nil, fmt.Errorf("%w: %s", astral.ErrBlueprintNotFound, objectType.String())
	}

	// decode the payload
	_, err = object.ReadFrom(bytes.NewReader(buf))
	switch {
	case err == nil:
		return

	case errors.Is(err, astral.ErrBlueprintNotFound) && b.AllowUnparsed:
		// if we're missing a blueprint, return an unparsed object if allowed
		return astral.NewUnparsedObject(objectType.String(), buf), nil

	default:
		return nil, err
	}

}

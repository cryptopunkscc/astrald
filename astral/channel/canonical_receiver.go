package channel

import (
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type CanonicalReceiver struct {
	r io.Reader
}

func NewCanonicalReceiver(r io.Reader) *CanonicalReceiver {
	return &CanonicalReceiver{r: r}
}

var _ Receiver = &CanonicalReceiver{}

func (r CanonicalReceiver) Receive() (object astral.Object, err error) {
	// read the stamp
	_, err = (&astral.Stamp{}).ReadFrom(r.r)
	if err != nil {
		return nil, fmt.Errorf("error reading stamp: %w", err)
	}

	// read the object type
	var objectType astral.ObjectType
	_, err = objectType.ReadFrom(r.r)
	if err != nil {
		return
	}

	object = astral.New(objectType.String())
	if object == nil {
		return nil, astral.NewErrBlueprintNotFound(objectType.String())
	}

	_, err = object.ReadFrom(r.r)
	if err != nil {
		return nil, fmt.Errorf("error reading object: %w", err)
	}

	return
}

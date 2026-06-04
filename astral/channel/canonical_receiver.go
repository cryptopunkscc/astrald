package channel

import (
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type CanonicalReceiver struct {
	r         io.Reader
	streamErr error
}

func NewCanonicalReceiver(r io.Reader) *CanonicalReceiver {
	return &CanonicalReceiver{r: r}
}

var _ Receiver = &CanonicalReceiver{}

// Canonical has no per-object framing: Stamp + tag + payload are read directly from the
// transport. Any error after Stamp consumption leaves the stream in an indeterminate state,
// so we latch the first non-nil error and refuse subsequent reads.
func (r *CanonicalReceiver) Receive() (object astral.Object, err error) {
	if r.streamErr != nil {
		return nil, r.streamErr
	}
	defer func() {
		if err != nil {
			r.streamErr = err
		}
	}()

	_, err = (&astral.Stamp{}).ReadFrom(r.r)
	if err != nil {
		return nil, fmt.Errorf("error reading stamp: %w", err)
	}

	var objectType astral.ObjectType
	_, err = objectType.ReadFrom(r.r)
	if err != nil {
		return
	}

	object = astral.New(objectType.String())
	if object == nil {
		return nil, fmt.Errorf("%w: %w: %s", astral.ErrStreamCorrupted, astral.ErrBlueprintNotFound, objectType.String())
	}

	_, err = object.ReadFrom(r.r)
	if err != nil {
		return nil, fmt.Errorf("error reading object: %w", err)
	}

	return
}

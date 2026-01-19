package channel

import (
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type CanonicalSender struct {
	w io.Writer
}

var _ Sender = &CanonicalSender{}

func NewCanonicalSender(w io.Writer) *CanonicalSender {
	return &CanonicalSender{w: w}
}

func (c CanonicalSender) Send(object astral.Object) (err error) {
	_, err = astral.Stamp{}.WriteTo(c.w)
	if err != nil {
		return fmt.Errorf("error writing stamp: %w", err)
	}

	_, err = astral.ObjectType(object.ObjectType()).WriteTo(c.w)
	if err != nil {
		return fmt.Errorf("error writing object type: %w", err)
	}

	_, err = object.WriteTo(c.w)
	if err != nil {
		return fmt.Errorf("error writing object: %w", err)
	}

	return
}

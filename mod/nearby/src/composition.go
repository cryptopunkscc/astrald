package nearby

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nearby"
)

const MaxAttachmentSize = 4 * 1024 //4kb

var _ nearby.Composition = &Composition{}

type Composition struct {
	receiver *astral.Identity
	s        *nearby.StatusMessage
}

func (c *Composition) Receiver() *astral.Identity {
	return c.receiver
}

func (c *Composition) Attach(o astral.Object) error {
	objectID, _ := astral.ResolveObjectID(o)
	if objectID.Size > MaxAttachmentSize {
		return errors.New("object too large")
	}
	return c.s.Attachments.Append(o)
}

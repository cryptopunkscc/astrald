package channel

import "github.com/cryptopunkscc/astrald/astral"

const (
	Binary    = "bin"
	JSON      = "json"
	Text      = "text"
	TextTyped = "text+"
	Render    = "render"
)

type ReceiveSender interface {
	Receiver
	Sender
}

type Receiver interface {
	Receive() (astral.Object, error)
}

type Sender interface {
	Send(astral.Object) error
}

package channel

import "github.com/cryptopunkscc/astrald/astral"

const (
	Canonical = "canonical"
	Binary    = "bin"
	JSON      = "json"
	Text      = "text"
	Render    = "render"
	Base64    = "base64"
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

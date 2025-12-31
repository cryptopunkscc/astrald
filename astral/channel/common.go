package channel

import "github.com/cryptopunkscc/astrald/astral"

const (
	Binary    = "bin"
	JSON      = "json"
	Text      = "text"
	TextTyped = "text+"
	Render    = "render"
)

type ReadWriter interface {
	Reader
	Writer
}

type Reader interface {
	Read() (astral.Object, error)
}

type Writer interface {
	Write(astral.Object) error
}

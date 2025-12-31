package channel

import "github.com/cryptopunkscc/astrald/astral"

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

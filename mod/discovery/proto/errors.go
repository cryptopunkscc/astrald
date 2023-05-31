package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/stream"
	"io"
)

var es stream.ErrorSpace

var (
	ErrRegistrationFailed = es.NewError(0xff, "registration failed")
)

func New(rw io.ReadWriter) *stream.Stream {
	return stream.NewStream(rw, es)
}

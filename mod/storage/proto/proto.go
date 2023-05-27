package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/stream"
	"io"
)

var es stream.ErrorSpace

var ErrUnavailable = es.NewError(0x01, "unavailable")

func New(rw io.ReadWriter) *stream.Stream {
	var s = stream.NewStream(rw, es)
	s.ErrorType = "c"
	return s
}

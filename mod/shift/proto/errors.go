package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/stream"
	"io"
)

var es stream.ErrorSpace

var (
	ErrRejected       = es.NewError(0x01, "rejected")
	ErrDenied         = es.NewError(0x02, "denied")
	ErrInvalidRequest = es.NewError(0xff, "invalid request")
)

func New(c io.ReadWriter) *stream.Stream {
	return stream.NewStream(c, es)
}

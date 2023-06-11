package proto

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"io"
)

type Session struct {
	*rpc.Session[string]
}

var es rpc.ErrorSpace

var (
	ErrRejected       = es.NewError(0x01, "rejected")
	ErrDenied         = es.NewError(0x02, "denied")
	ErrInvalidRequest = es.NewError(0xff, "invalid request")
)

func New(c io.ReadWriter) Session {
	return Session{rpc.NewSession[string](c, es)}
}

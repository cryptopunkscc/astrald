package rpc

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"io"
)

type Session struct {
	*rpc.Session[string]
}

var es rpc.ErrorSpace

var (
	ErrRegistrationFailed = es.NewError(0xff, "registration failed")
)

func New(rw io.ReadWriter) Session {
	return Session{rpc.NewSession[string](rw, es)}
}

func (s Session) Register(serviceName string) error {
	if err := s.Encode(MsgRegister{Service: serviceName}); err != nil {
		return err
	}
	return s.DecodeErr()
}

package proto

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
)

type Session struct {
	*rpc.Session[string]
}

var es rpc.ErrorSpace

var (
	ErrRejected        = es.NewError(0x01, "rejected")
	ErrDenied          = es.NewError(0x02, "denied")
	ErrUnableToProcess = es.NewError(0x03, "unable to process")
	ErrInvalidRequest  = es.NewError(0xff, "invalid request")
)

func New(c io.ReadWriter) Session {
	return Session{rpc.NewSession[string](c, es)}
}

func (s *Session) Shift(cert *RelayCert) error {
	if err := s.Encodef("[c]cv", CmdCert, cert); err != nil {
		return err
	}
	return s.DecodeErr()
}

func (s *Session) Query(target id.Identity, query string) error {
	err := s.Encodef("[c]cv", CmdQuery, proto.QueryParams{
		Identity: target,
		Query:    query,
	})
	if err != nil {
		return err
	}
	return s.DecodeErr()
}

package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
)

type Session struct {
	*cslq.Endec
	*proto.Conn
	ctx     context.Context
	mod     *Module
	appName string
}

func NewSession(mod *Module, conn net.Conn) *Session {
	return &Session{
		mod:   mod,
		Conn:  proto.NewConn(conn),
		Endec: cslq.NewEndec(conn),
	}
}

func (s *Session) auth(_ context.Context) error {
	var p proto.AuthParams
	if err := s.ReadMsg(&p); err != nil {
		return err
	}
	if len(p.Token) == 0 {
		if !s.mod.config.AllowAnonymous {
			s.WriteErr(proto.ErrUnauthorized)
			return errors.New("unauthorized")
		}
		s.appName = "anonymous"
	} else {
		app, ok := s.mod.tokens[p.Token]
		if !ok {
			return errors.New("unauthorized")
		}
		s.appName = app
	}

	return s.WriteErr(nil)
}

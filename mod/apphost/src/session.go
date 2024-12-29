package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
)

type Session struct {
	*cslq.Endec
	*proto.Conn
	ctx      context.Context
	mod      *Module
	remoteID *astral.Identity
	log      *log.Logger
}

func NewSession(mod *Module, conn net.Conn, log *log.Logger) *Session {
	return &Session{
		mod:   mod,
		Conn:  proto.NewConn(conn),
		Endec: cslq.NewEndec(conn),
		log:   log,
	}
}

func (s *Session) auth(_ context.Context) error {
	var p proto.AuthParams
	if err := s.ReadMsg(&p); err != nil {
		return err
	}

	if len(p.Token) > 0 {
		s.remoteID = s.mod.authToken(p.Token)
	}

	if s.remoteID.IsZero() {
		s.remoteID = s.mod.DefaultIdentity()
	}

	if s.remoteID.IsZero() {
		s.WriteErr(proto.ErrUnauthorized)
		return errors.New("unauthorized")
	}

	return s.WriteErr(nil)
}

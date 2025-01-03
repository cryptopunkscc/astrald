package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

func (s *Session) Serve(ctx context.Context) error {
	if err := s.auth(ctx); err != nil {
		return err
	}

	s.ctx = ctx
	defer s.Close()

	return cslq.Invoke(s, func(cmd proto.Command) error {
		switch cmd.Cmd {
		case proto.CmdQuery:
			return cslq.Invoke(s, s.query)

		case proto.CmdResolve:
			return cslq.Invoke(s, s.resolve)

		case proto.CmdRegister:
			return cslq.Invoke(s, s.register)

		case proto.CmdNodeInfo:
			return cslq.Invoke(s, s.nodeInfo)

		case proto.CmdExec:
			return cslq.Invoke(s, s.exec)

		default:
			return s.WriteErr(proto.ErrUnknownCommand)
		}
	})
}

func (s *Session) query(params proto.QueryParams) error {
	var err error
	var conn astral.Conn

	if params.Identity.IsZero() {
		params.Identity = s.remoteID
	}

	var q = astral.NewQuery(s.remoteID, params.Identity, params.Query)
	conn, err = query.Route(s.ctx, s.mod.node, q)

	if err == nil {
		s.WriteErr(nil)
		_, _, err := streams.Join(s, conn)
		return err
	}

	switch {
	case errors.Is(err, &astral.ErrRejected{}):
		return s.WriteErr(proto.ErrRejected)

	case errors.Is(err, &astral.ErrRouteNotFound{}):
		return s.WriteErr(proto.ErrRouteNotFound)

	default:
		s.mod.log.Error("unexpected error processing query: %v", err)
		return s.WriteErr(proto.ErrUnexpected)
	}
}

func (s *Session) resolve(p proto.ResolveParams) error {
	s.mod.log.Logv(2, "%v resolve %v", s.remoteID, p.Name)

	remoteID, err := s.mod.Dir.ResolveIdentity(p.Name)
	if err == nil {
		s.WriteErr(nil)
		return s.WriteMsg(proto.ResolveData{Identity: remoteID})
	}

	return s.WriteErr(proto.ErrFailed)
}

func (s *Session) nodeInfo(p proto.NodeInfoParams) error {
	s.mod.log.Logv(2, "%v nodeInfo %v", s.remoteID, p.Identity)

	s.WriteErr(nil)

	var data proto.NodeInfoData

	data.Identity = s.remoteID
	if !p.Identity.IsZero() {
		data.Identity = p.Identity
	}
	data.Name = s.mod.Dir.DisplayName(data.Identity)

	return s.WriteMsg(data)
}

func (s *Session) exec(params proto.ExecParams) error {
	var identity = s.remoteID

	if !params.Identity.IsZero() {
		identity = params.Identity
	}

	_, err := s.mod.Exec(identity, params.Exec, params.Args, params.Env)

	if err != nil {
		return s.WriteErr(proto.ErrFailed)
	}
	return s.WriteErr(nil)
}

func (s *Session) register(p proto.RegisterParams) error {
	s.mod.log.Logv(2, "%v register %v -> %v", s.remoteID, p.Service, p.Target)
	defer s.Close()

	// ...otherwise register under guest's router
	return s.registerGuest(p)
}

func (s *Session) registerGuest(p proto.RegisterParams) error {
	err := s.mod.addGuestRoute(s.remoteID, p.Service, p.Target)
	if err != nil {
		return s.WriteErr(proto.ErrAlreadyRegistered)
	}
	s.WriteErr(nil)

	// wait for the other party to close the session
	io.Copy(streams.NilWriter{}, s)

	return s.mod.removeGuestRoute(s.remoteID, p.Service)
}

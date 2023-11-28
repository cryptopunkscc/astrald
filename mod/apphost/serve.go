package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	routerpc "github.com/cryptopunkscc/astrald/mod/router/proto"
	"github.com/cryptopunkscc/astrald/net"
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
	var conn net.SecureConn

	if params.Identity.IsZero() {
		params.Identity = s.mod.node.Identity() //TODO: by default call your own service, not relay's
	}

	var query = net.NewQuery(s.remoteID, params.Identity, params.Query)
	conn, err = net.Route(s.ctx, s.mod.node.Router(), query)

	if err == nil {
		s.WriteErr(nil)
		_, _, err := streams.Join(s, conn)
		return err
	}

	switch {
	case errors.Is(err, net.ErrRejected),
		errors.Is(err, routerpc.ErrRejected):
		return s.WriteErr(proto.ErrRejected)

	case errors.Is(err, &net.ErrRouteNotFound{}):
		return s.WriteErr(proto.ErrRouteNotFound)

	case errors.Is(err, routerpc.ErrDenied):
		return s.WriteErr(proto.ErrUnauthorized)

	default:
		s.mod.log.Error("unexpected error processing query: %s", err)
		return s.WriteErr(proto.ErrUnexpected)
	}
}

func (s *Session) resolve(p proto.ResolveParams) error {
	s.mod.log.Logv(2, "%s resolve %s", s.remoteID, p.Name)

	remoteID, err := s.mod.node.Resolver().Resolve(p.Name)
	if err == nil {
		s.WriteErr(nil)
		return s.WriteMsg(proto.ResolveData{Identity: remoteID})
	}

	return s.WriteErr(proto.ErrFailed)
}

func (s *Session) nodeInfo(p proto.NodeInfoParams) error {
	s.mod.log.Logv(2, "%s nodeInfo %s", s.remoteID, p.Identity)

	s.WriteErr(nil)

	var data proto.NodeInfoData

	data.Identity = s.remoteID
	if !p.Identity.IsZero() {
		data.Identity = p.Identity
	}
	data.Name = s.mod.node.Resolver().DisplayName(data.Identity)

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
	s.mod.log.Logv(2, "%s register %s -> %s", s.remoteID, p.Service, p.Target)
	defer s.Close()

	err := s.mod.addGuestRoute(s.remoteID, p.Service, p.Target)
	if err != nil {
		return s.WriteErr(proto.ErrAlreadyRegistered)
	}
	s.WriteErr(nil)

	// wait for the other party to close the session
	io.Copy(streams.NilWriter{}, s)

	return s.mod.removeGuestRoute(s.remoteID, p.Service)
}

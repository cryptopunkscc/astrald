package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/streams"
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

		default:
			return s.WriteErr(proto.ErrUnknownCommand)
		}
	})
}

func (s *Session) query(p proto.QueryParams) error {
	if p.Identity.IsZero() {
		p.Identity = s.mod.node.Identity()
	}

	log.Logv(2, "<%s> query: %s %s", s.appName, p.Identity, p.Query)
	conn, err := s.mod.node.Query(context.Background(), p.Identity, p.Query)

	if err == nil {
		s.WriteErr(nil)
		_, _, err := streams.Join(s, conn)
		return err
	}

	switch {
	case errors.Is(err, services.ErrRejected),
		errors.Is(err, link.ErrRejected),
		errors.Is(err, services.ErrServiceNotFound):
		return s.WriteErr(proto.ErrRejected)

	case errors.Is(err, services.ErrTimeout):
		return s.WriteErr(proto.ErrTimeout)

	default:
		log.Error("query: %s", err)
		return s.WriteErr(proto.ErrUnexpected)
	}
}

func (s *Session) resolve(p proto.ResolveParams) error {
	log.Logv(2, "<%s> resolve: %s", s.appName, p.Name)

	remoteID, err := s.mod.node.Resolver().Resolve(p.Name)
	if err == nil {
		s.WriteErr(nil)
		return s.WriteMsg(proto.ResolveData{Identity: remoteID})
	}

	return s.WriteErr(proto.ErrFailed)
}

func (s *Session) nodeInfo(p proto.NodeInfoParams) error {
	log.Logv(2, "<%s> nodeInfo: %s", s.appName, p.Identity)

	s.WriteErr(nil)

	var data proto.NodeInfoData

	if p.Identity.IsZero() {
		data.Identity = s.mod.node.Identity()
		data.Name = s.mod.node.Alias()
	} else {
		data.Identity = p.Identity
		data.Name = s.mod.node.Resolver().DisplayName(p.Identity)
	}

	return s.WriteMsg(data)
}

func (s *Session) register(p proto.RegisterParams) error {
	log.Logv(2, "<%s> register: %s -> %s", s.appName, p.Service, p.Target)

	ctx, cancel := context.WithCancel(context.Background())
	srv, err := s.mod.node.Services().RegisterContext(ctx, p.Service)
	if err == nil {
		s.WriteErr(nil)
		go func() {
			var buf [16]byte
			for {
				if _, err := s.Read(buf[:]); err != nil {
					break
				}
			}
			cancel()
			s.Close()
		}()
		s.forwardService(srv, p.Target)
		return nil
	}
	cancel()

	switch {
	case errors.Is(err, services.ErrAlreadyRegistered):
		return s.WriteErr(proto.ErrAlreadyRegistered)

	default:
		return s.WriteErr(proto.ErrUnexpected)
	}
}

func (s *Session) forwardService(srv *services.Service, target string) {
	for query := range srv.Queries() {
		c, err := proto.Dial(target)
		if err != nil {
			log.Errorv(2, "%s -> %s dial error: %s", srv.Name(), target, err)
			query.Reject()
			continue
		}

		conn := proto.NewConn(c)

		err = conn.WriteMsg(proto.InQueryParams{
			Identity: query.RemoteIdentity(),
			Query:    query.Query(),
		})
		if err != nil {
			continue
		}

		if conn.ReadErr() != nil {
			query.Reject()
			conn.Close()
			continue
		}

		qConn, err := query.Accept()
		if err != nil {
			conn.Close()
			continue
		}

		go streams.Join(conn, qConn)
	}
}

package apphost

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"github.com/cryptopunkscc/astrald/mod/shift"
	shiftp "github.com/cryptopunkscc/astrald/mod/shift/proto"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"sync"
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
	if params.Identity.IsZero() {
		params.Identity = s.mod.node.Identity()
	}

	s.mod.log.Logv(2, "%s query %s:%s", s.remoteID, params.Identity, params.Query)

	var err error
	var conn io.ReadWriteCloser

	if params.Identity.IsEqual(s.mod.node.Identity()) {
		conn, err = s.mod.node.Services().Query(s.ctx, s.remoteID, params.Query, nil)
	} else {
		var caller id.Identity

		if s.remoteID.IsEqual(s.mod.node.Identity()) {
			conn, err = s.mod.node.Network().Query(s.ctx, params.Identity, params.Query)
		} else {
			// Get private key from the store
			caller, err = s.mod.keys.Find(s.remoteID)
			if err != nil {
				s.mod.log.Errorv(2, "no private key for %s", s.remoteID)
				return proto.ErrUnauthorized
			}

			conn, err = s.mod.node.Network().Query(s.ctx, params.Identity, shift.ServiceName)
			if err != nil {
				s.mod.log.Errorv(2, "error shifting identity: %s", err)
				return proto.ErrUnauthorized
			}
			defer conn.Close()

			c := shiftp.New(conn)

			var shiftParams shiftp.ShiftParams

			shiftParams, err = shiftp.BuildShiftParams(caller, s.mod.node.Identity())
			if err != nil {
				return err
			}

			err = c.Encode(
				shiftp.Cmd{Cmd: shiftp.CmdShift},
				shiftParams,
			)
			if err != nil {
				return err
			}
			if err = c.DecodeErr(); err != nil {
				return err
			}

			err = c.Encode(shiftp.Cmd{Cmd: shiftp.CmdQuery},
				shiftp.QueryParams{
					Query: params.Query,
				})
			if err != nil {
				return err
			}

			err = c.DecodeErr()
		}
	}

	if err == nil {
		s.WriteErr(nil)
		_, _, err := streams.Join(s, conn)
		return err
	}

	switch {
	case errors.Is(err, services.ErrRejected),
		errors.Is(err, link.ErrRejected),
		errors.Is(err, shiftp.ErrRejected),
		errors.Is(err, services.ErrServiceNotFound):
		return s.WriteErr(proto.ErrRejected)

	case errors.Is(err, shiftp.ErrDenied):
		return s.WriteErr(proto.ErrUnauthorized)

	case errors.Is(err, services.ErrTimeout):
		return s.WriteErr(proto.ErrTimeout)

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

	if p.Identity.IsZero() {
		data.Identity = s.mod.node.Identity()
		data.Name = s.mod.node.Alias()
	} else {
		data.Identity = p.Identity
		data.Name = s.mod.node.Resolver().DisplayName(p.Identity)
	}

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

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	defer s.Close()

	var queries = services.NewQueryChan(1)

	service, err := s.mod.node.Services().Register(ctx, s.remoteID, p.Service, queries.Push)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrAlreadyRegistered):
			return s.WriteErr(proto.ErrAlreadyRegistered)

		default:
			return s.WriteErr(proto.ErrUnexpected)
		}
	}
	s.WriteErr(nil)

	go func() {
		io.Copy(streams.NilWriter{}, s)
		service.Close()
	}()

	go func() {
		<-service.Done()
		close(queries)
	}()

	var wg sync.WaitGroup
	wg.Add(128)
	for i := 0; i < 128; i++ {
		go func() {
			s.forwardWorker(queries, p.Target)
			wg.Done()
		}()
	}
	wg.Wait()

	return nil
}

func (s *Session) forwardWorker(queries services.QueryChan, target string) {
	for query := range queries {
		c, err := proto.Dial(target)
		if err != nil {
			s.mod.log.Errorv(2, "%s -> %s dial error: %s", query.Query(), target, err)
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

		streams.Join(conn, qConn)
	}
}

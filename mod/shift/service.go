package shift

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shift/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"io"
)

const ServiceName = "sys.shift"

var _ query.Router = &Service{}

type Service struct {
	*Module
}

func (service *Service) Run(ctx context.Context) error {
	_, err := service.node.Services().Register(ctx, service.node.Identity(), ServiceName, service)

	return err
}

func (service *Service) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
		if err := service.startSession(ctx, conn, q.Origin()); err != nil {

			service.log.Logv(2, "(%s) session err: %s", q.Caller(), err)
		}
	})
}

func (service *Service) startSession(ctx context.Context, conn net.SecureConn, origin string) error {
	var err error
	var identity = conn.RemoteIdentity()
	var c = proto.New(conn)
	defer c.Close()

	for {
		var cmd proto.Cmd
		err = c.Decode(&cmd)
		if err != nil {
			return err
		}

		switch cmd.Cmd {
		case proto.CmdShift:
			var params proto.ShiftParams
			if err := c.Decode(&params); err != nil {
				return err
			}
			service.log.Logv(2, "(%s) shifting to %s...", identity, params.Identity)

			if !params.Verify(identity) {
				if err := c.EncodeErr(proto.ErrDenied); err != nil {
					return err
				}
				continue
			}

			identity = params.Identity

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

		case proto.CmdQuery:
			var params proto.QueryParams
			if err := c.Decode(&params); err != nil {
				return err
			}

			service.log.Logv(2, "(%s) executing query %s...", identity, params.Query)

			query := query.NewOrigin(identity, id.Identity{}, params.Query, origin)

			localWriter, err := service.node.Services().RouteQuery(ctx, query, conn)
			if err != nil {
				return c.EncodeErr(proto.ErrRejected)
			}

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

			io.Copy(localWriter, conn)

			return nil

		default:
			return c.EncodeErr(proto.ErrInvalidRequest)
		}
	}
}

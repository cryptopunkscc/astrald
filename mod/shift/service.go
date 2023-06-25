package shift

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shift/proto"
	"github.com/cryptopunkscc/astrald/net"
	"io"
)

const ServiceName = "sys.shift"

var _ net.Router = &Service{}

type Service struct {
	*Module
}

func (service *Service) Run(ctx context.Context) error {
	_, err := service.node.Services().Register(ctx, service.node.Identity(), ServiceName, service)

	return err
}

func (service *Service) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		if err := service.serve(ctx, conn, query.Origin()); err != nil {
			service.log.Errorv(2, "(%s) serve: %s", query.Caller(), err)
		}
	})
}

func (service *Service) serve(ctx context.Context, conn net.SecureConn, origin string) error {
	var err error
	var caller = conn.RemoteIdentity()
	var c = proto.New(conn)
	defer c.Close()

	service.log.Logv(2, "(%s) connected", caller)

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

			if !params.Verify(caller) {
				service.log.Logv(2, "(%s) shift to %s denied", caller, params.Identity)
				if err := c.EncodeErr(proto.ErrDenied); err != nil {
					return err
				}
				continue
			}

			service.log.Logv(2, "(%s) shifted to %s", caller, params.Identity)

			caller = params.Identity

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

		case proto.CmdQuery:
			var params proto.QueryParams
			if err := c.Decode(&params); err != nil {
				return err
			}

			var q = net.NewOrigin(caller, id.Identity{}, params.Query, origin)
			var shiftedConn = replaceIdentity{SecureConn: conn, remoteIdentity: caller}

			service.log.Logv(2, "(%s) routing query -> %s:%s", caller, q.Target(), q.Query())

			localWriter, err := service.node.Services().RouteQuery(ctx, q,
				shiftedConn)
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

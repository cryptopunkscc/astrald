package shift

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/shift/proto"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/streams"
)

const ServiceName = "sys.shift"

type Service struct {
	*Module
}

func (srv *Service) Run(ctx context.Context) error {
	_, err := srv.node.Services().Register(ctx, srv.node.Identity(), ServiceName, srv.onQuery)

	return err
}

func (srv *Service) onQuery(ctx context.Context, query *services.Query) error {
	conn, err := query.Accept()
	if err == nil {
		go func() {
			err := srv.startSession(ctx, conn, query.Link())
			srv.log.Logv(2, "(%s) session err: %s", query.RemoteIdentity(), err)
		}()
	}

	return err
}

func (srv *Service) startSession(ctx context.Context, conn *services.Conn, origin *link.Link) error {
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
			srv.log.Logv(2, "(%s) shifting to %s...", identity, params.Identity)

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

			srv.log.Logv(2, "(%s) executing query %s...", identity, params.Query)

			q, err := srv.node.Services().Query(ctx, identity, params.Query, origin)

			if err != nil {
				return c.EncodeErr(proto.ErrRejected)
			}

			if err := c.EncodeErr(nil); err != nil {
				return err
			}

			streams.Join(q, conn)

			return nil

		default:
			return c.EncodeErr(proto.ErrInvalidRequest)
		}
	}
}

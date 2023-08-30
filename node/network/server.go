package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
)

type SecureConnHandlerFunc func(conn net.SecureConn) error
type LinkHandlerFunc func(net.Link) error

type Server struct {
	localID  id.Identity
	listener infra.Listener
	handler  LinkHandlerFunc
	log      *log.Logger
}

func newServer(localID id.Identity, i infra.Infra, handler LinkHandlerFunc, log *log.Logger) (*Server, error) {
	listener, ok := i.(infra.Listener)
	if !ok {
		return nil, errors.New("infra is not a listener")
	}

	srv := &Server{
		localID:  localID,
		listener: listener,
		handler:  handler,
		log:      log,
	}

	return srv, nil
}

func (srv *Server) Run(ctx context.Context) error {
	listenCh, err := srv.listener.Listen(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case conn, ok := <-listenCh:
			if !ok {
				return nil
			}

			l, err := link.Accept(ctx, conn, srv.localID)

			switch {
			case err == nil:
				srv.handler(l)

			case errors.Is(err, io.EOF),
				errors.Is(err, context.DeadlineExceeded),
				errors.Is(err, context.Canceled):
				srv.log.Errorv(2, "inbound handshake error: %s", err)

			default:
				srv.log.Errorv(1, "inbound handshake error: %s", err)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

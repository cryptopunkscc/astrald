package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"io"
)

type SecureConnHandlerFunc func(conn net.SecureConn) error

type Server struct {
	localID  id.Identity
	listener infra.Listener
	handler  SecureConnHandlerFunc
	log      *log.Logger
}

func newServer(localID id.Identity, listener infra.Listener, handler SecureConnHandlerFunc, log *log.Logger) (*Server, error) {
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

			authConn, err := srv.Handshake(ctx, conn)
			switch {
			case err == nil:
				srv.handler(authConn)

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

func (srv *Server) Handshake(ctx context.Context, conn net.Conn) (secureConn net.SecureConn, err error) {
	hsCtx, _ := context.WithTimeout(ctx, HandshakeTimeout)
	secureConn, err = auth.HandshakeInbound(hsCtx, conn, srv.localID)

	if err != nil {
		conn.Close()
	}

	return
}

func (srv *Server) SetHandler(handler SecureConnHandlerFunc) {
	srv.handler = handler
}

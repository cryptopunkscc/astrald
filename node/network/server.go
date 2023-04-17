package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
)

type AuthConnHandlerFunc func(auth.Conn) error

type Server struct {
	localID  id.Identity
	listener infra.Listener
	handler  AuthConnHandlerFunc
}

func newServer(localID id.Identity, listener infra.Listener, handler AuthConnHandlerFunc) (*Server, error) {
	srv := &Server{
		localID:  localID,
		listener: listener,
		handler:  handler,
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
				log.Errorv(2, "inbound handshake error: %s", err)

			default:
				log.Errorv(1, "inbound handshake error: %s", err)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (srv *Server) Handshake(ctx context.Context, conn infra.Conn) (authConn auth.Conn, err error) {
	hsCtx, _ := context.WithTimeout(ctx, HandshakeTimeout)
	authConn, err = auth.HandshakeInbound(hsCtx, conn, srv.localID)

	if err != nil {
		conn.Close()
	}

	return
}

func (srv *Server) SetHandler(handler AuthConnHandlerFunc) {
	srv.handler = handler
}

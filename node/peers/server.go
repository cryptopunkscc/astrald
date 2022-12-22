package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link"
	"io"
	"log"
	"time"
)

const HandshakeTimeout = 15 * time.Second

type Server struct {
	localID  id.Identity
	listener infra.Listener
	conns    chan infra.Conn
	err      error
	links    *link.Link
}

func newServer(localID id.Identity, listener infra.Listener) (*Server, error) {
	srv := &Server{
		localID:  localID,
		listener: listener,
		conns:    make(chan infra.Conn, 8),
	}

	return srv, nil
}

func (srv *Server) AddConn(conn infra.Conn) {
	srv.conns <- conn
}

func (srv *Server) Run(ctx context.Context) (chan *link.Link, error) {
	listenCh, err := srv.listener.Listen(ctx)
	if err != nil {
		return nil, err
	}

	ch := make(chan *link.Link)

	go func() {
		for conn := range listenCh {
			srv.conns <- conn
		}
		close(srv.conns)
		defer close(ch)
	}()

	go func() {
		for conn := range srv.conns {
			link, err := srv.process(ctx, conn)
			switch {
			case err == nil:
				ch <- link
			case errors.Is(err, io.EOF):
			case errors.Is(err, context.DeadlineExceeded):
			case errors.Is(err, context.Canceled):
			default:
				log.Println("peers.Server: inbound handshake error:", err)
			}

		}
	}()

	return ch, nil
}

func (srv *Server) process(ctx context.Context, conn infra.Conn) (*link.Link, error) {
	hctx, _ := context.WithTimeout(ctx, HandshakeTimeout)
	auth, err := auth.HandshakeInbound(hctx, conn, srv.localID)

	if err != nil {
		conn.Close()
		return nil, err

	}

	return link.New(auth), nil
}

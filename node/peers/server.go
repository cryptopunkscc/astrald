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
	Conns chan infra.Conn
	links chan *link.Link
}

func runServer(ctx context.Context, localID id.Identity, listener infra.Listener) (*Server, error) {
	listenCh, err := listener.Listen(ctx)
	if err != nil {
		return nil, err
	}

	srv := &Server{
		Conns: make(chan infra.Conn),
		links: make(chan *link.Link),
	}

	go srv.process(ctx, localID)

	go func() {
		defer close(srv.links)
		for conn := range listenCh {
			srv.Conns <- conn
		}
	}()

	return srv, nil
}

func (i *Server) process(ctx context.Context, localID id.Identity) {
	for conn := range i.Conns {
		conn := conn
		go func() {
			hctx, _ := context.WithTimeout(ctx, HandshakeTimeout)
			auth, err := auth.HandshakeInbound(hctx, conn, localID)
			if err != nil {
				switch {
				case errors.Is(err, io.EOF):
				case errors.Is(err, context.DeadlineExceeded):
				case errors.Is(err, context.Canceled):
				default:
					log.Println("peers.Server.process(): inbound handshake error:", err)
				}
				conn.Close()
				return
			}
			i.links <- link.New(auth)
		}()
	}
}

func (i *Server) Links() chan *link.Link {
	return i.links
}

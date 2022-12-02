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
)

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
			auth, err := auth.HandshakeInbound(ctx, conn, localID)
			if (err != nil) && (!errors.Is(err, io.EOF)) {
				log.Println("peers.Server.process(): inbound handshake error:", err)
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

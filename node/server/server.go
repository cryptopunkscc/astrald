package server

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Server struct {
	Conns chan infra.Conn
	links chan *link.Link
}

func Run(ctx context.Context, localID id.Identity, listener infra.Listener) (*Server, error) {
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
		auth, err := auth.HandshakeInbound(ctx, conn, localID)
		if err != nil {
			conn.Close()
			continue
		}
		i.links <- link.New(auth)
	}
}

func (i Server) Links() chan *link.Link {
	return i.links
}

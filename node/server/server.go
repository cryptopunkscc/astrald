package server

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Server struct {
	links chan *link.Link
	Conns chan infra.Conn
}

func Run(ctx context.Context, localID id.Identity) (*Server, error) {
	listenCh, _ := astral.Listen(ctx)

	srv := &Server{
		links: make(chan *link.Link),
		Conns: make(chan infra.Conn),
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

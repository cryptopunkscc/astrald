package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	_net "net"
	"strings"
)

var _ Server = &TCPServer{}

type TCPServer struct {
	*Module
	bind     string
	target   net.Router
	listener _net.Listener
}

func NewTCPServer(mod *Module, bind string, target net.Router) (*TCPServer, error) {
	var err error
	var srv = &TCPServer{
		Module: mod,
		target: target,
		bind:   bind,
	}

	srv.listener, err = _net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (srv *TCPServer) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		srv.listener.Close()
	}()

	for {
		client, err := srv.listener.Accept()
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "use of closed network connection"):
				return nil
			default:
				return err
			}
		}

		go func() {
			var query = net.NewQuery(id.Identity{}, id.Identity{}, "")
			var src = net.NewSecurePipeWriter(client, srv.node.Identity())

			dst, err := srv.target.RouteQuery(ctx, query, src, net.DefaultHints())
			if err != nil {
				client.Close()
				return
			}
			defer dst.Close()

			io.Copy(dst, client)
		}()
	}
}

func (srv *TCPServer) Target() net.Router {
	return srv.target
}

func (srv *TCPServer) String() string {
	return "tcp://" + srv.bind
}

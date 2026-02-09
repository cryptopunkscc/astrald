package tcp

import (
	"context"
	"net"
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Server struct {
	*Module
}

func NewServer(module *Module) *Server {
	return &Server{Module: module}
}

func (srv *Server) Run(ctx context.Context) error {
	// start the listener
	var addrStr = ":" + strconv.Itoa(srv.config.ListenPort)

	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		srv.log.Errorv(0, "failed to start server: %v", err)
		return err
	}

	endpoint, _ := tcp.ParseEndpoint(listener.Addr().String())

	srv.log.Info("started server at %v", endpoint)
	defer srv.log.Info("stopped server at %v", endpoint)

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	// accept connections
	for {
		rawConn, err := listener.Accept()
		if err != nil {
			return err
		}

		var conn = tcp.WrapConn(rawConn, false)

		go func() {
			err := srv.Nodes.AcceptInboundLink(ctx, conn)
			if err != nil {
				srv.log.Errorv(1, "handshake failed from %v: %v", conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

func (mod *Module) startServer(ctx context.Context) {
	srv := NewServer(mod)
	if err := srv.Run(astral.NewContext(ctx)); err != nil {
		mod.log.Errorv(1, "server error: %v", err)
	}
}

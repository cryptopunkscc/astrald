package tcp

import (
	"context"
	_net "net"
	"strconv"
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

	listener, err := _net.Listen("tcp", addrStr)
	if err != nil {
		srv.log.Errorv(0, "failed to start server: %v", err)
		return err
	}

	endpoint, _ := Parse(listener.Addr().String())

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

		var conn = wrapTCPConn(rawConn, false)

		go func() {
			err := srv.nodes.Accept(ctx, conn)
			if err != nil {
				srv.log.Errorv(1, "handshake failed from %v: %v", conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

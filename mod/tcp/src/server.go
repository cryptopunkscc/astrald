package tcp

import (
	"net"
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Server struct {
	*Module
	listenPort int
}

func NewServer(module *Module, listenPort int) *Server {
	return &Server{
		Module:     module,
		listenPort: listenPort,
	}
}

func (s *Server) Run(ctx *astral.Context) error {
	// start the listener
	var addrStr = ":" + strconv.Itoa(s.listenPort)

	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		s.log.Errorv(0, "failed to start server: %v", err)
		return err
	}

	endpoint, _ := tcp.ParseEndpoint(listener.Addr().String())

	s.log.Info("started server at %v", endpoint)
	defer s.log.Info("stopped server at %v", endpoint)

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
			err := s.Nodes.Accept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "handshake failed from %v: %v", conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

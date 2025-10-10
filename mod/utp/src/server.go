package utp

import (
	"fmt"
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	utpmod "github.com/cryptopunkscc/astrald/mod/utp"
	"github.com/cryptopunkscc/utp"
)

// Server implements UDP listening with connection acceptance via rudp.Listener
type Server struct {
	*Module
	listener *utp.Listener
	acceptCh chan *utp.Conn
}

// NewServer creates a new src UDP server
func NewServer(module *Module) *Server {
	return &Server{
		Module:   module,
		acceptCh: make(chan *utp.Conn, 16),
	}
}

// Run starts the server and listens for incoming connections
func (s *Server) Run(ctx *astral.Context) error {
	addr := &net.UDPAddr{Port: s.config.ListenPort}
	listenerAddr, err := utp.ResolveAddr("utp", addr.String())
	if err != nil {
		return fmt.Errorf(`utp server/run: resolve address %v: %w`, addr, err)
	}

	utpListener, err := utp.Listen("utp", listenerAddr)
	if err != nil {
		return fmt.Errorf(`utp server/run: failed to start utp listener at %v: %w`, addr, err)
	}

	s.listener = utpListener
	localEndpoint, err := utpmod.ParseEndpoint(utpListener.Addr().String())
	if err != nil {
		return fmt.Errorf(`utp server/run: failed to parse local endpoint %v: %w`, utpListener.Addr(), err)
	}

	s.log.Info("started server at %v", utpListener.Addr())
	go func() {
		select {
		case <-ctx.Done():
			// will cancel listener
			s.log.Info("stopped server at %v", utpListener.Addr())
			_ = utpListener.Close()
			return
		}
	}()

	for {
		utpConn, err := utpListener.AcceptUTP()
		if err != nil {
			return fmt.Errorf(`utp server/run: failed to accept utp connection: %w`, err)
		}

		remoteEndpoint, _ := utpmod.ParseEndpoint(utpConn.RemoteAddr().String())
		s.log.Info("accepted connection from %v", remoteEndpoint)

		var conn = WrapUtpConn(utpConn, localEndpoint, remoteEndpoint, false)

		go func() {
			err := s.Nodes.Accept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "utp server/run: handshake failed from %v: %v",
					conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	return nil
}

package udp

import (
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/udp/rudp"
)

// Server implements UDP listening with connection acceptance via rudp.Listener
type Server struct {
	*Module
	rListener *rudp.Listener
	acceptCh  chan *rudp.Conn
}

// NewServer creates a new src UDP server
func NewServer(module *Module) *Server {
	return &Server{
		Module:   module,
		acceptCh: make(chan *rudp.Conn, 16),
	}
}

// Run starts the server and listens for incoming connections
func (s *Server) Run(ctx *astral.Context) error {

	addr := &net.UDPAddr{Port: s.config.ListenPort}
	hto := s.config.DialTimeout
	if hto <= 0 {
		hto = 5 * time.Second
	}
	rListener, err := rudp.Listen(ctx, addr, s.Module.config.TransportConfig, hto)
	if err != nil {
		s.log.Errorv(0, "failed to start rudp listener: %v", err)
		return err
	}
	s.rListener = rListener
	s.log.Info("started server at %v", rListener.Addr())
	defer s.log.Info("stopped server at %v", rListener.Addr())

	// Accept loop
	go func() {
		acceptCtx := ctx
		for {
			conn, err := rListener.Accept(acceptCtx)
			if err != nil {
				return
			}
			select {
			case s.acceptCh <- conn:
			default:
				// drop if application not consuming fast enough
			}
		}
	}()

	<-ctx.Done()
	return s.Close()
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	if s.rListener != nil {
		_ = s.rListener.Close()
	}
	return nil
}

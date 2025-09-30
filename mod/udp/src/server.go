package udp

import (
	"net"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

// Server implements src UDP listener with connection demultiplexing
type Server struct {
	*Module
	listener *net.UDPConn
	conns    map[string]*Conn // Remote address → connection map
	mutex    sync.Mutex       // Protects access to conns
	acceptCh chan *Conn       // Channel for accepted connections
	stopCh   chan struct{}    // Channel to signal server shutdown
	wg       sync.WaitGroup   // WaitGroup for managing goroutines
}

// NewServer creates a new src UDP server
func NewServer(module *Module) *Server {
	return &Server{
		Module:   module,
		conns:    make(map[string]*Conn),
		acceptCh: make(chan *Conn, 16),
		stopCh:   make(chan struct{}),
	}
}

// Run starts the server and listens for incoming connections
func (s *Server) Run(ctx *astral.Context) error {

	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: s.config.ListenPort})
	if err != nil {
		s.log.Errorv(0, "failed to start server: %v", err)
		return err
	}
	s.listener = listener
	s.log.Info("started server at %v", listener.LocalAddr())
	defer s.log.Info("stopped server at %v", listener.LocalAddr())

	localEndpoint, err := udp.ParseEndpoint(listener.LocalAddr().
		String())
	if err != nil {
		s.log.Errorv(1, "error parsing local endpoint: %v", err)
		return err
	}

	s.wg.Add(1)
	go s.readLoop(ctx, localEndpoint)

	<-ctx.Done()
	s.Close()
	return nil
}

// Close gracefully shuts down the server
func (s *Server) Close() error {
	close(s.stopCh)
	s.mutex.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.mutex.Unlock()

	s.listener.Close()
	s.wg.Wait()
	return nil
}

func (s *Server) readLoop(ctx *astral.Context, localEndpoint *udp.Endpoint) {
	defer s.wg.Done()

	buf := make([]byte, 64*1024) // TODO: Max packet size?
	for {
		n, addr, err := s.listener.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-s.stopCh:
				return // Graceful shutdown
			default:
				s.log.Errorv(1, "read error: %v", err)
				continue
			}
		}

		pkt := &Packet{}
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			s.log.Errorv(1, "packet unmarshal error from %v: %v", addr, err)
			continue // drop malformed
		}

		remoteKey := addr.String()
		s.mutex.Lock()
		conn, foundConn := s.conns[remoteKey]
		if !foundConn && pkt.Flags&FlagSYN != 0 {
			remoteEndpoint, err := udp.ParseEndpoint(addr.String())
			if err != nil {
				s.log.Errorv(1, "ParseEndpoint error for %v: %v", addr, err)
				continue
			}

			conn, err = NewConn(s.listener, localEndpoint, remoteEndpoint, s.Module.config.TransportConfig)
			if err != nil {
				s.log.Errorv(1, "NewConn error for %v: %v", addr, err)
				s.mutex.Unlock()
				continue
			}

			conn.outbound = false // mark as inbound connection

			conn.inCh = make(chan *Packet, 128)
			s.conns[remoteKey] = conn
			go func() {
				err := conn.StartServerHandshake(ctx, pkt)
				if err != nil {
					s.log.Errorv(1, "handshake error for %v: %v", addr, err)
				}
			}()
		}
		s.mutex.Unlock()

		if conn != nil {
			select {
			case conn.inCh <- pkt:
				// success
			default:
				s.log.Errorv(1, "inCh full for %v, dropping packet", addr)
			}
		}
	}
}

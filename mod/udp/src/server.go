package udp

import (
	"net"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

// Connection states
const (
	StateClosed      = iota // Connection is closed
	StateSynSent            // SYN sent, waiting for SYN-ACK
	StateSynReceived        // SYN received, waiting for ACK
	StateEstablished        // Connection established
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

	s.wg.Add(1)
	go s.readLoop()

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

// readLoop handles incoming datagrams and routes them to connections
func (s *Server) readLoop() {
	defer s.wg.Done()
	buf := make([]byte, 64*1024) // Large buffer for high throughput

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

		s.handlePacket(buf[:n], addr)
	}
}

// handlePacket processes an incoming packet and routes it to the appropriate connection
func (s *Server) handlePacket(data []byte, addr *net.UDPAddr) {
}

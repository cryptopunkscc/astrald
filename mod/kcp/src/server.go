package kcp

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

// Server implements KCP listening with connection acceptance via kcp.Listener
type Server struct {
	*Module
	listener   *kcpgo.Listener
	listenPort int
	onAccept   exonet.EphemeralHandler
	closed     atomic.Bool
}

func NewServer(module *Module, listenPort int, onAccept exonet.EphemeralHandler) *Server {
	return &Server{
		Module:     module,
		listenPort: listenPort,
		onAccept:   onAccept,
	}
}

func (s *Server) Run(ctx *astral.Context) error {
	listener, err := NewKcpListener(s.listenPort)
	if err != nil {
		return err
	}

	kcpListener, ok := listener.(*kcpgo.Listener)
	if !ok {
		return fmt.Errorf("kcp server/run: failed to cast listener to kcp listener")
	}
	s.listener = kcpListener

	localEndpoint, err := kcpmod.ParseEndpoint(kcpListener.Addr().String())
	if err != nil {
		return fmt.Errorf(`kcp server/run: failed to parse local endpoint %v: %w`,
			kcpListener.Addr(), err)
	}

	s.log.Info("started server at %v", kcpListener.Addr())

	// Shutdown goroutine
	go func() {
		<-ctx.Done()
		s.Close() // triggers clean shutdown
	}()

	for {
		// Block accept until ctx signals shutdown or listener closes
		sess, err := kcpListener.AcceptKCP()
		if err != nil {
			// Expected shutdown (listener.Close was called)
			if s.closed.Load() || ctx.Err() != nil {
				s.log.Info("stopped server at %v", kcpListener.Addr())
				return nil
			}

			// Unexpected accept error
			return fmt.Errorf("kcp server/run: accept failed: %w", err)
		}

		remoteEndpoint, _ := kcpmod.ParseEndpoint(sess.RemoteAddr().String())
		s.log.Info("accepted connection from %v", remoteEndpoint)

		conn := WrapKCPConn(sess, localEndpoint, remoteEndpoint, false)

		// Accept handler
		go func() {
			shouldClose, err := s.onAccept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "kcp server/onAccept error from %v: %v",
					conn.RemoteEndpoint(), err)
				return
			}

			if shouldClose {
				s.Close()
			}
		}()
	}
}

func (s *Server) Close() error {
	if s.closed.Swap(true) {
		return nil // already closed
	}

	if s.listener != nil {
		_ = s.listener.Close() // this unblocks AcceptKCP()
	}

	return nil
}

func NewKcpListener(listenPort int) (net.Listener, error) {
	addr := fmt.Sprintf(":%d", listenPort)
	kcpListener, err := kcpgo.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to start kcp listener at %v: %w", addr, err)
	}
	return kcpListener, nil
}

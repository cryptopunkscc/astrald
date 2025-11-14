package kcp

import (
	"fmt"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

var _ exonet.EphemeralListener = &Server{}

// Server implements KCP listening with connection acceptance via kcp.Listener
type Server struct {
	*Module
	listenPort astral.Uint16
	listener   *kcpgo.Listener
	onAccept   exonet.EphemeralHandler
	closed     atomic.Bool
	closedCh   chan struct{}
}

func NewServer(module *Module, listenPort astral.Uint16, onAccept exonet.EphemeralHandler) *Server {
	return &Server{
		Module:     module,
		listenPort: listenPort,
		onAccept:   onAccept,
		closedCh:   make(chan struct{}),
	}
}

func (s *Server) Run(ctx *astral.Context) error {
	addr := fmt.Sprintf(":%d", s.listenPort)
	kcpListener, err := kcpgo.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		return fmt.Errorf("kcp server/run: failed to listen on %v: %w", addr, err)
	}

	s.listener = kcpListener

	localEndpoint, err := kcpmod.ParseEndpoint(kcpListener.Addr().String())
	if err != nil {
		return fmt.Errorf(`kcp server/run: failed to parse local endpoint %v: %w`,
			kcpListener.Addr(), err)
	}

	s.log.Info("started server at %v", kcpListener.Addr())
	go func() {
		select {
		case <-ctx.Done():
			s.Close()
		case <-s.Done():
		}
	}()

	for {
		sess, err := kcpListener.AcceptKCP()
		if err != nil {
			if s.closed.Load() || ctx.Err() != nil {
				s.log.Info("stopped server at %v", kcpListener.Addr())
				return nil
			}

			return fmt.Errorf("kcp server/run: accept failed: %w", err)
		}

		remoteEndpoint, _ := kcpmod.ParseEndpoint(sess.RemoteAddr().String())
		s.log.Info("accepted connection from %v", remoteEndpoint)

		conn := WrapKCPConn(sess, localEndpoint, remoteEndpoint, false)
		go func() {
			shouldClose, err := s.onAccept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "kcp server/onAccept error from %v: %v", conn.RemoteEndpoint(), err)
				return
			}

			if shouldClose {
				s.Close()
			}
		}()
	}
}

func (s *Server) Done() <-chan struct{} {
	return s.closedCh
}

func (s *Server) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}

	if s.listener != nil {
		return s.listener.Close()
	}

	close(s.closedCh)
	return nil
}

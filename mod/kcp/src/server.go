package kcp

import (
	"fmt"
	"net"

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
	onAccept   exonet.AcceptHandler
}

func NewServer(module *Module, listenPort int, onAccept exonet.AcceptHandler) *Server {
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
		return fmt.Errorf(`kcp server/run: failed to parse local endpoint %v: %w`, kcpListener.Addr(), err)
	}

	s.log.Info("started server at %v", kcpListener.Addr())
	go func() {
		<-ctx.Done()
		s.log.Info("stopped server at %v", kcpListener.Addr())
		_ = kcpListener.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		sess, err := kcpListener.AcceptKCP()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf(`kcp server/run: failed to accept kcp connection: %w`, err)
		}

		remoteEndpoint, _ := kcpmod.ParseEndpoint(sess.RemoteAddr().String())
		s.log.Info("accepted connection from %v", remoteEndpoint)

		var conn = WrapKCPConn(sess, localEndpoint, remoteEndpoint, false)

		go func() {
			err := s.onAccept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "kcp server/run: onAccept failed from %v: %v",
					conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

func (s *Server) Close() error {
	if s.listener != nil {
		_ = s.listener.Close()
	}
	return nil
}

func NewKcpListener(listenPort int) (listener net.Listener, err error) {
	addr := fmt.Sprintf(":%d", listenPort)
	kcpListener, err := kcpgo.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		return nil, fmt.Errorf(`failed to start kcp listener at %v: %w`, addr, err)
	}

	return kcpListener, nil
}

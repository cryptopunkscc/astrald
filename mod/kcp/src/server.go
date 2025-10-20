package kcp

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	kcpmod "github.com/cryptopunkscc/astrald/mod/kcp"
	kcpgo "github.com/xtaci/kcp-go/v5"
)

// Server implements KCP listening with connection acceptance via kcp.Listener
type Server struct {
	*Module
	listener *kcpgo.Listener
}

func NewServer(module *Module) *Server {
	return &Server{Module: module}
}

func (s *Server) Run(ctx *astral.Context) error {
	addr := fmt.Sprintf(":%d", s.config.ListenPort)
	kcpListener, err := kcpgo.ListenWithOptions(addr, nil, 0, 0)
	if err != nil {
		return fmt.Errorf(`kcp server/run: failed to start kcp listener at %v: %w`, addr, err)
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
			err := s.Nodes.Accept(ctx, conn)
			if err != nil {
				s.log.Errorv(1, "kcp server/run: handshake failed from %v: %v",
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

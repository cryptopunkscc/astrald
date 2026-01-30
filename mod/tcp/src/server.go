package tcp

import (
	"net"
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Server struct {
	*Module
}

func NewServer(module *Module) *Server {
	return &Server{Module: module}
}

func (srv *Server) Run(ctx *astral.Context) error {
	// start the listener
	var addrStr = ":" + strconv.Itoa(srv.config.ListenPort)

	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		srv.log.Errorv(0, "failed to start server: %v", err)
		return err
	}

	endpoint, _ := tcp.ParseEndpoint(listener.Addr().String())

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

		var conn = tcp.WrapConn(rawConn, false)

		go func() {
			err := srv.Nodes.Accept(ctx, conn)
			if err != nil {
				srv.log.Errorv(1, "handshake failed from %v: %v", conn.RemoteEndpoint(), err)
				return
			}
		}()
	}
}

func (mod *Module) startServer() {
	if mod.ctx == nil {
		return
	}

	mod.serverMu.Lock()
	defer mod.serverMu.Unlock()

	if mod.serverCancel != nil {
		return // already running
	}

	ctx, cancel := mod.ctx.WithCancel()
	mod.serverCancel = cancel

	go func() {
		srv := NewServer(mod)
		if err := srv.Run(ctx); err != nil {
			mod.log.Errorv(1, "server error: %v", err)
		}
	}()
}

func (mod *Module) stopServer() {
	mod.serverMu.Lock()
	defer mod.serverMu.Unlock()

	if mod.serverCancel == nil {
		return // not running
	}

	mod.serverCancel()
	mod.serverCancel = nil
}

func (mod *Module) switchServer(enable *astral.Bool) {
	if enable == nil || *enable {
		mod.startServer()
	} else {
		mod.stopServer()
	}
}

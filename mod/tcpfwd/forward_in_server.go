package tcpfwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	_net "net"
	"strings"
	"time"
)

type ForwardInServer struct {
	*Module
	identity id.Identity
	tcpAddr  string
	target   string
}

func (server *ForwardInServer) Run(ctx context.Context) error {
	listener, err := _net.Listen("tcp", server.tcpAddr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	server.log.Logv(1, "forwarding %s to %s", server.tcpAddr, server.target)

	for {
		inConn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			if err := server.serve(inConn); err != nil {
				server.log.Errorv(1, "serve error: %v", err)
			}
		}()
	}
}

func (server *ForwardInServer) serve(in _net.Conn) error {
	defer in.Close()

	var err error
	var nodeHex, queryText string
	var caller = server.node.Identity()

	t := server.target
	if idx := strings.Index(t, "@"); idx != -1 {
		callerName := t[:idx]
		t = t[idx+1:]

		caller, err = server.node.Resolver().Resolve(callerName)
		if err != nil {
			return err
		}

		keystore, err := server.assets.KeyStore()
		if err != nil {
			return err
		}

		caller, err = keystore.Find(caller)
		if err != nil {
			return err
		}
	}

	var parts = strings.SplitN(t, ":", 2)
	if len(parts) == 2 {
		nodeHex, queryText = parts[0], parts[1]
	} else {
		nodeHex, queryText = "localnode", parts[0]
	}

	target, err := server.node.Resolver().Resolve(nodeHex)
	if err != nil {
		return err
	}

	query := net.NewQuery(caller, target, queryText)
	src := net.NewSecurePipeWriter(in, caller)

	ctx, cancel := context.WithTimeout(server.ctx, 30*time.Second)
	defer cancel()

	dst, err := server.node.Router().RouteQuery(ctx, query, src, net.DefaultHints())
	if err != nil {
		return err
	}

	io.Copy(dst, in)
	dst.Close()

	return nil
}

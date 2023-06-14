package tcpfwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/streams"
	_net "net"
	"strings"
)

type ForwardInServer struct {
	*Module
	tcpAddr string
	target  string
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

		go server.serve(inConn)
	}
}

func (server *ForwardInServer) serve(in _net.Conn) {
	var nodeHex, q string

	var parts = strings.SplitN(server.target, ":", 2)
	if len(parts) == 2 {
		nodeHex, q = parts[0], parts[1]
	} else {
		nodeHex, q = "localnode", parts[0]
	}

	nodeID, err := server.node.Resolver().Resolve(nodeHex)
	if err != nil {
		return
	}

	out, err := query.Run(server.ctx, server.node, query.New(server.node.Identity(), nodeID, q))
	if err != nil {
		in.Close()
		return
	}

	streams.Join(in, out)
}

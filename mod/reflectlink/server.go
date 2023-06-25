package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Server struct {
	*Module
}

func (server *Server) Run(ctx context.Context) error {
	s, err := server.node.Services().Register(ctx, server.node.Identity(), serviceName, server)
	if err != nil {
		return err
	}

	disco, err := modules.Find[*discovery.Module](server.node.Modules())
	if err == nil {
		disco.AddSourceContext(ctx, server, server.node.Identity())
	}

	<-s.Done()

	return nil
}

func (server *Server) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if linker, ok := caller.(net.Linker); ok {
		if l := linker.Link(); l != nil {
			return net.Accept(query, caller, func(conn net.SecureConn) {
				server.reflectLink(conn, l)
			})
		}
	}

	return nil, net.ErrRejected
}

func (server *Server) reflectLink(conn net.SecureConn, sourceLink net.Link) error {
	defer conn.Close()

	var e = sourceLink.Transport().RemoteEndpoint()
	var ref proto.Reflection

	if e != nil {
		ref.RemoteEndpoint = proto.Endpoint{
			Network: e.Network(),
			Address: e.String(),
		}
	}

	return json.NewEncoder(conn).Encode(ref)
}

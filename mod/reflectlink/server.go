package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/query"
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

func (server *Server) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if linker, ok := remoteWriter.(query.Linker); ok {
		if l, ok := linker.Link().(*link.Link); ok {
			return query.Accept(q, remoteWriter, func(conn net.SecureConn) {
				server.reflectLink(conn, l)
			})
		}
	}

	return nil, link.ErrRejected
}

func (server *Server) reflectLink(conn net.SecureConn, sourceLink *link.Link) error {
	defer conn.Close()

	var e = sourceLink.RemoteEndpoint()
	var ref proto.Reflection

	ref.RemoteEndpoint = proto.Endpoint{
		Network: e.Network(),
		Address: e.String(),
	}

	return json.NewEncoder(conn).Encode(ref)
}

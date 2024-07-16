package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/net"
	"reflect"
)

type Server struct {
	*Module
}

func (server *Server) Run(ctx context.Context) error {
	err := server.node.LocalRouter().AddRoute(serviceName, server)
	if err != nil {
		return err
	}
	defer server.node.LocalRouter().RemoveRoute(serviceName)

	if server.sdp != nil {
		server.sdp.AddServiceDiscoverer(server)
		defer server.sdp.RemoveServiceDiscoverer(server)
	}

	<-ctx.Done()

	return nil
}

func (server *Server) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// Reject queries coming from sources without transport
	var output = net.FinalOutput(caller)
	var t, ok = output.(exonet.Transporter)
	if !ok {
		return net.Reject()
	}
	if t.Transport() == nil {
		return net.Reject()
	}

	return net.Accept(query, caller, server.reflect)
}

func (server *Server) reflect(conn net.Conn) {
	defer conn.Close()

	var t, _ = net.FinalOutput(conn).(exonet.Transporter)
	var output, ok = t.Transport().(exonet.Conn)
	if !ok {
		return
	}
	var endpoint = output.RemoteEndpoint()

	if endpoint == nil {
		server.log.Errorv(2, "link with %v has no remote endpoint (transport type %v)",
			conn.RemoteIdentity(),
			reflect.TypeOf(t.Transport()),
		)
		return
	}

	var reflection proto.Reflection

	if endpoint != nil {
		reflection.RemoteEndpoint = proto.Endpoint{
			Network: endpoint.Network(),
			Address: endpoint.Address(),
		}
	}

	json.NewEncoder(conn).Encode(reflection)

	server.log.Infov(2, "reflected %v", conn.RemoteIdentity())

	return
}

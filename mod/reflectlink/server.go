package reflectlink

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/reflectlink/proto"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
)

type Server struct {
	*Module
}

func (mod *Server) Run(ctx context.Context) error {
	_, err := mod.node.Services().Register(ctx, mod.node.Identity(), serviceName, mod.handleQuery)
	if err != nil {
		return err
	}

	disco, err := modules.Find[*discovery.Module](mod.node.Modules())
	if err == nil {
		disco.AddSource(mod, mod.node.Identity())
		go func() {
			<-ctx.Done()
			disco.RemoveSource(mod)
		}()
	}

	<-ctx.Done()

	return nil
}

func (mod *Server) handleQuery(ctx context.Context, query *services.Query) error {
	mod.log.Logv(2, "query from %s", query.RemoteIdentity())
	if query.Origin() == services.OriginLocal {
		query.Reject()
		return nil
	}

	conn, err := query.Accept()
	if err != nil {
		return nil
	}

	if err := mod.reflect(ctx, conn); err != nil {
		mod.log.Error("reflect: %s", err)
	}

	return nil
}

func (mod *Server) reflect(ctx context.Context, conn *services.Conn) error {
	defer conn.Close()

	var e = conn.Link().RemoteEndpoint()
	var ref proto.Reflection

	ref.RemoteEndpoint = proto.Endpoint{
		Network: e.Network(),
		Address: e.String(),
	}

	return json.NewEncoder(conn).Encode(ref)
}

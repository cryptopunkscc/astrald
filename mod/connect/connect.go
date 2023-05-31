package connect

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
)

const serviceName = "connect"

type Connect struct {
	node node.Node
}

func (mod *Connect) Run(ctx context.Context) error {
	_, err := mod.node.Services().Register(ctx, mod.node.Identity(), serviceName, func(query *services.Query) error {
		mod.handleQuery(ctx, query)
		return nil
	})
	if err != nil {
		return err
	}

	<-ctx.Done()

	return nil
}

func (mod *Connect) handleQuery(ctx context.Context, query *services.Query) error {
	// skip local queries
	if query.Source() == services.SourceLocal {
		query.Reject()
		return errors.New("local query not allowed")
	}

	conn, err := query.Accept()
	if err != nil {
		return err
	}

	infraConn := &wrapper{
		local:           mod.node.Identity().Public(),
		remote:          query.RemoteIdentity(),
		ReadWriteCloser: conn,
		outbound:        false,
	}

	authConn, err := mod.node.Network().Server().Handshake(ctx, infraConn)
	if err != nil {
		infraConn.Close()
		return err
	}

	return mod.node.Network().AddSecureConn(authConn)
}

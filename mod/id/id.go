package id

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	_node "github.com/cryptopunkscc/astrald/node"
)

const serviceHandle = "id"

type Id struct{}

func (i Id) Run(ctx context.Context, node *_node.Node) error {
	port, err := node.Ports.Register(serviceHandle)
	if err != nil {
		return err
	}
	defer port.Close()

	go func() {
		for query := range port.Queries() {
			conn := query.Accept()

			enc.WriteIdentity(conn, node.Identity())
			conn.Close()
		}
	}()

	<-ctx.Done()
	return nil
}

func (i Id) String() string {
	return serviceHandle
}

func Query() (identity id.Identity, err error) {
	conn, err := astral.Query("", "id")
	if err != nil {
		return
	}
	identity, err = enc.ReadIdentity(conn)
	if err != nil {
		return
	}
	return
}

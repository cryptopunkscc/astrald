package id

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/legacy/enc"
	"github.com/cryptopunkscc/astrald/lib/astral"
	_node "github.com/cryptopunkscc/astrald/node"
	"log"
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
			conn, err := query.Accept()
			if err != nil {
				log.Println("Cannot accept query", err)
				continue
			}

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
	conn, err := astral.Dial(id.Identity{}, "id")
	if err != nil {
		return
	}
	identity, err = enc.ReadIdentity(conn)
	if err != nil {
		return
	}
	return
}

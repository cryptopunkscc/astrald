package apphost

import (
	"context"
	_node "github.com/cryptopunkscc/astrald/node"
	"log"
	"net"
)

type AppHost struct{}

const ModuleName = "apphost"

func (AppHost) Run(ctx context.Context, node *_node.Node) error {
	for conn := range Serve(ctx) {
		go func(conn net.Conn) {
			client := NewClient(conn, node)

			err := client.handle(ctx)
			if err != nil {
				log.Println("(apphost) client error:", err)
			}
		}(conn)
	}
	return nil
}

func (AppHost) String() string {
	return ModuleName
}

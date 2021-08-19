package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"net"
)

func main(ctx context.Context, core api.Core) error {
	for conn := range Serve(ctx) {
		go func(conn net.Conn) {
			client := NewClient(conn, core.Network())

			err := client.handle(ctx)
			if err != nil {
				log.Println("apps: client error:", err)
			}
		}(conn)
	}
	return nil
}

func init() {
	_ = node.RegisterService("apps", main)
}

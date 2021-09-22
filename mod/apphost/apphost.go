package apphost

import (
	"context"
	_node "github.com/cryptopunkscc/astrald/node"
	"log"
	"net"
)

func main(ctx context.Context, node *_node.Node) error {
	for conn := range Serve(ctx) {
		go func(conn net.Conn) {
			client := NewClient(conn, node)

			err := client.handle(ctx)
			if err != nil {
				log.Println("apphost: client error:", err)
			}
		}(conn)
	}
	return nil
}

func init() {
	_ = _node.RegisterService("apphost", main)
}

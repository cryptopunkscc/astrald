package astralmobile

import (
	"context"
	_ "github.com/cryptopunkscc/astrald/mod/admin"
	_ "github.com/cryptopunkscc/astrald/mod/apphost"
	_node "github.com/cryptopunkscc/astrald/node"
	"log"
)

var identity string
var stop context.CancelFunc

func Start(astralHome string) error {
	log.Println("Staring astrald")

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	stop = shutdown
	node, err := _node.Run(ctx, astralHome)
	if err != nil {
		panic(err)
	}

	identity = node.Identity().String()

	<-ctx.Done()

	// Run the node
	return nil
}

func Identity() string {
	return identity
}

func Stop() {
	stop()
}

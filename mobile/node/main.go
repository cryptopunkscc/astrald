package astralmobile

import (
	"context"
	_node "github.com/cryptopunkscc/astrald/node"
	_ "github.com/cryptopunkscc/astrald/mod/admin"
	_ "github.com/cryptopunkscc/astrald/mod/apphost"
	"log"
)

var identity string
var stop context.CancelFunc

func Start(astralHome string) error {
	log.Println("Staring astrald")

	// Instantiate the node
	node := _node.New(astralHome)

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	stop = shutdown
	identity = node.Identity.String()

	// Run the node
	return node.Run(ctx)
}

func Identity() string {
	return identity
}

func Stop() {
	stop()
}

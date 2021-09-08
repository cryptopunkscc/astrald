package astralandroid

import (
	"context"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services"
	_ "github.com/cryptopunkscc/astrald/services/apphost"
	_ "github.com/cryptopunkscc/astrald/services/warpdrive/init"
	"log"
)

var identity string
var stop context.CancelFunc

func Start(astralHome string) error {
	log.Println("Staring astrald")

	// Instantiate the node
	node := _node.New(astralHome)

	// init AstralHome for service
	services.AstralHome = astralHome

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

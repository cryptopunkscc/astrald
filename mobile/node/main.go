package astralmobile

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/node"
	"log"
)

var identity string
var stop context.CancelFunc

func Start(astralHome string, btAdapter BTAdapter) error {
	if btAdapter != nil {
		bt.Default = newBTWrapper(btAdapter)
	}

	log.Println("Staring astrald")

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	stop = shutdown
	n, err := node.Run(
		ctx,
		astralHome,
		admin.Admin{},
		apphost.AppHost{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
	)
	if err != nil {
		panic(err)
	}

	identity = n.Identity().String()

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

package astralandroid

import (
	"context"
	"github.com/cryptopunkscc/astrald/bind/api"
	_node "github.com/cryptopunkscc/astrald/node"
	_ "github.com/cryptopunkscc/astrald/services/apphost"
	"log"
	"time"
)

var stop context.CancelFunc

func Start(astralHome string) {
	// Figure out the config path
	log.Println("log Staring astrald")

	// Instantiate the node
	node := _node.New(astralHome)

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	stop = shutdown

	// Run the node
	err := node.Run(ctx)

	time.Sleep(50 * time.Millisecond)

	// Check results
	if err != nil {
		log.Printf("error: %s\n", err)
	}
}

func Register(
	name string,
	service astralApi.Service,
) error {
	return _node.RegisterService(
		name,
		serviceRunner(service),
	)
}

func Stop() {
	stop()
}

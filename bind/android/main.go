package astralandroid

import (
	"context"
	"github.com/cryptopunkscc/astrald/bind/api"
	_node "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services"
	_ "github.com/cryptopunkscc/astrald/services/apphost"
	_ "github.com/cryptopunkscc/astrald/services/identifier/init"
	_ "github.com/cryptopunkscc/astrald/services/identity/init"
	_ "github.com/cryptopunkscc/astrald/services/lore/init"
	_ "github.com/cryptopunkscc/astrald/services/messenger/init"
	_ "github.com/cryptopunkscc/astrald/services/push/init"
	_ "github.com/cryptopunkscc/astrald/services/repo/init"
	_ "github.com/cryptopunkscc/astrald/services/share/init"
	_ "github.com/cryptopunkscc/astrald/services/warpdrive/init"
	"log"
	"time"
)

// Exit statuses
const (
	ExitSuccess   = int(iota) // Normal exit
	ExitNodeError             // Node reported an error
)

var stop context.CancelFunc

func Start(astralHome string) {
	// Figure out the config path
	log.Println("log Staring astrald")

	// init AstralHome for file system service
	services.AstralHome = astralHome

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

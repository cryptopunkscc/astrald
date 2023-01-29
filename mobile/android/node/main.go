package astral

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/optimizer"
	"github.com/cryptopunkscc/astrald/mod/reflectlink"
	"github.com/cryptopunkscc/astrald/mod/roam"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"os"
	"path/filepath"
	"time"
)

var identity string
var stop context.CancelFunc

func Start(
	dir string,
	handlers Handlers,
	bluetooth Bluetooth,
) error {
	astralRoot := filepath.Join(dir, "node")
	err := os.MkdirAll(astralRoot, 0700)
	if err != nil {
		return err
	}

	if bluetooth != nil {
		bt.Instance = newBluetoothAdapter(bluetooth)
	}

	log.Println("Staring astrald")
	astral.ListenProtocol = "tcp"

	// Set up app execution context
	var ctx context.Context
	ctx, stop = context.WithCancel(context.Background())

	// start the node
	n, err := node.New(
		astralRoot,
		admin.Loader{},
		apphost.Loader{},
		connect.Loader{},
		gateway.Loader{},
		reflectlink.Loader{},
		roam.Loader{},
		contacts.Loader{},
		optimizer.Loader{},
		handlerLoader("android", handlers),
	)
	if err != nil {
		fmt.Println("init error:", err)
		return err
	}

	// Run the node
	go func() {
		if err := n.Run(ctx); err != nil {
			fmt.Printf("run error: %s\n", err)
		}
	}()

	identity = n.Identity().String()

	<-ctx.Done()

	time.Sleep(300 * time.Millisecond)

	log.Println("Astral stopped")

	return nil
}

func Identity() string {
	return identity
}

func Stop() {
	stop()
}

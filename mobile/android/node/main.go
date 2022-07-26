package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"log"
	"os"
	"path/filepath"
)

var identity string
var stop context.CancelFunc
var ctx context.Context
var dataDir string

func Start(
	dir string,
	mods Modules,
	bluetooth Bluetooth,
) error {
	dataDir = dir
	nodeDir := filepath.Join(dataDir, "node")
	err := os.MkdirAll(nodeDir, 0700)
	if err != nil {
		return err
	}

	if bluetooth != nil {
		bt.Instance = newBluetoothAdapter(bluetooth)
	}

	log.Println("Staring astrald")
	astral.ListenProtocol = "tcp"

	// Set up app execution context
	ctx, stop = context.WithCancel(context.Background())

	m := []node.ModuleRunner{
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
	}
	m = append(m, androidRunners(mods)...)

	n, err := node.Run(ctx, nodeDir, m...)

	if err != nil {
		return err
	}

	identity = n.Identity().String()

	<-ctx.Done()

	return nil
}

func Identity() string {
	return identity
}

func Stop() {
	stop()
}

type Writer io.Writer

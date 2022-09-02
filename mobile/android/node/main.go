package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/core"
	warpdrive "github.com/cryptopunkscc/astrald/cmd/warpdrived/server"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var identity string
var stop context.CancelFunc

func Start(
	dir string,
	handlers Handlers,
	bluetooth Bluetooth,
) error {
	nodeDir := filepath.Join(dir, "node")
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
	var ctx context.Context
	ctx, stop = context.WithCancel(context.Background())
	services := &sync.WaitGroup{}

	m := []node.ModuleRunner{
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
		contacts.Contacts{},
		handlerRunner("android", handlers),
		serviceRunner(services, "warpdrive",
			&warpdrive.Server{
				Debug: true,
				Component: core.Component{
					Config: core.Config{
						Platform:       core.PlatformAndroid,
						RepositoryDir:  filepath.Join(dir, "warpdrive"),
						RemoteResolver: true,
					},
				},
			},
		),
	}

	n, err := node.Run(ctx, nodeDir, m...)

	if err != nil {
		return err
	}

	identity = n.Identity().String()

	<-ctx.Done()

	services.Wait()

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

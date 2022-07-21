package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/app/warpdrive"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/mobile/android/node/content"
	"github.com/cryptopunkscc/astrald/mobile/android/node/notify"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
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
var astralNode *node.Node
var dataDir string

func Start(
	dir string,
	bluetooth Bluetooth,
	api AndroidApi,
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

	adapter := androidApi{api: api}
	n, err := node.Run(
		ctx, nodeDir,
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
		contacts.Contacts{},
		notify.CreateChannel{Api: adapter},
		notify.DispatchNotification{Api: adapter},
		content.GetInfo{Api: adapter},
		content.Read{Api: adapter},
	)
	if err != nil {
		return err
	}
	astralNode = n

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

func StartWarpdrive() {
	warpdrive.Service{
		Context: ctx,
		Api: &astralApi{
			ctx:  ctx,
			node: astralNode,
		},
		Core: api.Core{
			Config: api.Config{
				Platform:       api.PlatformAndroid,
				RepositoryDir:  filepath.Join(dataDir, "warpdrive"),
				RemoteResolver: true,
			},
		},
	}.Run()
}

type Writer io.Writer

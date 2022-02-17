package astralmobile

import (
	"context"
	"github.com/cryptopunkscc/astrald/app/warpdrive"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/infra/bt"
	content "github.com/cryptopunkscc/astrald/mobile/android/service/content/go"
	notify "github.com/cryptopunkscc/astrald/mobile/android/service/notification/go"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/id"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/node"
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
	btNetwork BTClient,
	nativeNotifier NativeAndroidNotify,
	nativeContentResolver NativeAndroidContentResolver,
) error {
	dataDir = dir
	nodeDir := filepath.Join(dataDir, "node")
	err := os.MkdirAll(nodeDir, 0700)
	if err != nil {
		return err
	}

	if btNetwork != nil {
		bt.Instance = newBTAdapter(btNetwork)
	}

	notifier := &AndroidNotify{nativeNotifier}
	contentResolver := &AndroidContentResolver{nativeContentResolver}

	log.Println("Staring astrald")
	astral.Instance().UseTCP = true

	// Set up app execution context
	c, shutdown := context.WithCancel(context.Background())
	ctx = c

	stop = shutdown
	n, err := node.Run(
		ctx, nodeDir,
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
		id.Id{},
		contacts.Contacts{},
		notify.CreateChannel{Api: notifier},
		notify.DispatchNotification{Api: notifier},
		content.GetInfo{Api: contentResolver},
		content.Read{Api: contentResolver},
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
		Api:     newApiAdapter(ctx, astralNode),
		Core: api.Core{
			Config: api.Config{
				Platform:       api.PlatformAndroid,
				RepositoryDir:  filepath.Join(dataDir, "warpdrive"),
				RemoteResolver: true,
			},
		},
	}.Run()
}

package setup

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/android"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/desktop"
	"github.com/cryptopunkscc/astrald/app/warpdrive/platform/stub"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"log"
	"os"
	"path/filepath"
)

func Core(core *api.Core) {
	// Defaults
	core.Logger = log.Default()
	core.Sys = &api.Sys{}
	core.Channel = &api.Channel{}
	core.Cache = &api.Cache{
		Mutex:    &api.Mutex{},
		Incoming: api.Offers{},
		Outgoing: api.Offers{},
		Peers:    api.Peers{},
	}
	core.Observers = &api.Observers{
		IncomingOffers: api.NewSubscriptions(),
		IncomingStatus: api.NewSubscriptions(),
		OutgoingOffers: api.NewSubscriptions(),
		OutgoingStatus: api.NewSubscriptions(),
	}

	// Platform
	if core.Platform == "" {
		core.Platform = api.PlatformDefault
	}

	// Storage
	core.StorageDir = storageDir()

	// Repository
	if core.RepositoryDir == "" {
		core.RepositoryDir = repositoryDir()
	}
	file.Outgoing(*core).Init()
	file.Incoming(*core).Init()

	// Peers
	core.Cache.Peers = file.Peers(*core).Get()

	// Offers
	core.Cache.Incoming = file.Incoming(*core).Get()
	core.Cache.Outgoing = file.Outgoing(*core).Get()

	// Notify
	core.Sys.Notify = newNotify(*core)
}

func storageDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, "warpdrive", "received")
	return dir
}

func repositoryDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(0)
	}
	dir := filepath.Join(cfgDir, "warpdrive")
	os.MkdirAll(dir, 0700)
	return dir
}

func newNotify(core api.Core) (notify api.Notify) {
	switch core.Platform {
	case api.PlatformDesktop:
		notify = desktop.Notify
	case api.PlatformAndroid:
		notifier := &android.Notifier{}
		notifier = notifier.Init()
		if notifier != nil {
			notify = notifier.Notify
		}
	default:
		notify = stub.Notify
	}
	return
}

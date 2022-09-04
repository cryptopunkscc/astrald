package warpdrived

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/core"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/notify/android"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/notify/stub"
	"github.com/cryptopunkscc/astrald/lib/warpdrived/storage/file"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func setupCore(c *core.Component) {
	// Defaults
	c.Logger = log.Default()
	c.Sys = &core.Sys{}
	c.Channel = &core.Channel{}
	c.Cache = &core.Cache{
		Mutex:    &core.Mutex{},
		Incoming: warpdrive.Offers{},
		Outgoing: warpdrive.Offers{},
		Peers:    warpdrive.Peers{},
	}
	c.Observers = &core.Observers{
		IncomingOffers: warpdrive.NewSubscriptions(),
		IncomingStatus: warpdrive.NewSubscriptions(),
		OutgoingOffers: warpdrive.NewSubscriptions(),
		OutgoingStatus: warpdrive.NewSubscriptions(),
	}

	// Platform
	if c.Platform == "" {
		c.Platform = core.PlatformDefault
	}

	// Storage
	c.StorageDir = storageDir()

	// Repository
	if c.RepositoryDir == "" {
		c.RepositoryDir = repositoryDir()
	}

	// Peers
	c.Cache.Peers = file.Peers(*c).Get()

	// Offers
	c.Cache.Incoming = file.Incoming(*c).Get()
	c.Cache.Outgoing = file.Outgoing(*c).Get()

	// Notify
	switch c.Platform {
	case core.PlatformAndroid:
		c.Sys.Notify = android.New(c.Api)
	default:
		c.Sys.Notify = stub.Notify
	}

	// Workers
	c.Job = &sync.WaitGroup{}
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

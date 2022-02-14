package service

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/storage/file"
	"log"
	"os"
	"path/filepath"
)

func Core(config api.Config) api.Core {
	init := initializer{}
	init.Config = config
	init.core()
	init.platform()
	init.storage()
	init.repository()
	init.offers()
	init.peers()
	init.notifier()
	return api.Core(init)
}

type initializer api.Core

func (core *initializer) core() {
	core.Logger = log.Default()
	core.Notifications = &api.Notifications{}
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
}

func (core *initializer) platform() {
	if core.Platform == "" {
		core.Platform = api.PlatformDefault
	}
}

func (core *initializer) storage() {
	core.StorageDir = storageDir()
}

func (core *initializer) repository() {
	if core.RepositoryDir == "" {
		core.RepositoryDir = repositoryDir()
	}
	c := api.Core(*core)
	file.Outgoing(c).Init()
	file.Incoming(c).Init()
}

func (core *initializer) peers() {
	core.Cache.Peers = file.Peers(*core).Get()
}

func (core *initializer) offers() {
	c := api.Core(*core)
	core.Cache.Incoming = file.Incoming(c).Get()
	core.Cache.Outgoing = file.Outgoing(c).Get()
}

func (core *initializer) notifier() {
	notifier := Notify{}
	notifier.Core = (*api.Core)(core)
	notifier.Init()
	go notifier.Start()
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

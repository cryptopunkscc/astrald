package core

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/local"
	"github.com/cryptopunkscc/astrald/app/warpdrive/remote"
	"os"
	"path/filepath"
)

func (c *core) Setup() {
	c.setupCore()
	c.setupStorage()
	c.setupResolver()
	c.setupRepository()
	c.setupOffers()
	c.setupPeers()
}

func (c *core) setupCore() {
	c.persistence = &persistence{}
	c.cache = &cache{}
	c.peers = api.Peers{}
	c.incoming = api.Offers{}
	c.outgoing = api.Offers{}
	c.observers = &observers{}
	c.filesOffers = api.NewSubscriptions()
	c.incomingStatus = api.NewSubscriptions()
	c.outgoingStatus = api.NewSubscriptions()
}

func (c *core) setupResolver() {
	if c.RemoteResolver {
		c.Resolver = remote.NewResolver()
	} else {
		c.Resolver = local.NewResolver()
	}
}

func (c *core) setupStorage() {
	c.Storage = local.NewStorage(storageDir())
}

func (c *core) setupRepository() {
	if c.RepositoryDir == "" {
		c.RepositoryDir = repositoryDir()
	}
	c.Repository = local.NewRepository(c.RepositoryDir)
}

func (c *core) setPeers(peers api.Peers) {
	c.peers = peers
}

func (c *core) setupPeers() {
	peers := c.Peers().List()
	for _, peer := range peers {
		peerRef := peer
		c.peers[peer.Id] = &peerRef
	}
}

func (c *core) setupOffers() {
	c.incoming = c.Repository.Incoming().List()
	c.outgoing = c.Repository.Outgoing().List()
}

func storageDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	dir := filepath.Join(home, "warpdrive", "received")
	err = os.MkdirAll(dir, 0755)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}
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

package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/hub"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/infra"
	nlink "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/linking"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/node/server"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/storage"
	"log"
)

// Run starts the node, waits for it to finish and returns an error if any
func Run(ctx context.Context, dataDir string, modules ...ModuleRunner) (*Node, error) {
	var err error

	// Storage
	store := storage.NewFilesystemStorage(dataDir)

	// Config
	cfg, err := config.Load(store)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	node := &Node{
		Config:  cfg,
		Store:   store,
		Ports:   hub.New(),
		Peers:   peer.NewManager(),
		peers:   make(map[string]*peer.Peer),
		queries: make(chan *nlink.Query),
		events:  &sig.Queue{},
		links:   make(chan *alink.Link, 1),
	}

	// Set up identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("identity setup error: %w", err)
	}

	// Say hello
	nodeKey := node.identity.PublicKeyHex()
	if node.Alias() != "" {
		nodeKey = fmt.Sprintf("%s (%s)", node.Alias(), nodeKey)
	}
	log.Printf("astral node %s statrting...", nodeKey)

	// Set up the infrastructure
	node.Infra, err = infra.New(
		node.Config.Infra,
		infra.FilteredQuerier{Querier: node, FilteredID: node.identity},
		node.Store,
	)
	if err != nil {
		log.Println("infra error:", err)
		return nil, err
	}

	// Contacts
	node.Contacts = contacts.New(store)

	// Server
	node.Server, err = server.Run(ctx, node.identity, node.Infra)
	if err != nil {
		return nil, err
	}

	// Modules
	node.runModules(ctx, modules)

	// Presence
	node.Presence = presence.Run(ctx, node.Infra)

	node.Linking = linking.Run(ctx, node.identity, node.Contacts, node.Peers, node.Infra)

	err = node.Presence.Announce(ctx, node.identity)
	if err != nil {
		log.Println("announce error:", err) // non-critical
	}

	// Run it
	go node.process(ctx)

	return node, nil
}

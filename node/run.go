package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
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
		Config: cfg,
		Store:  store,
	}

	// Set up identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("identity setup error: %w", err)
	}

	// Set up local hub
	node.Ports = hub.New(&node.events)

	// Say hello
	nodeKey := node.identity.PublicKeyHex()
	if node.Alias() != "" {
		nodeKey = fmt.Sprintf("%s (%s)", node.Alias(), nodeKey)
	}
	log.Printf("astral node %s statrting...", nodeKey)

	// Set up the infrastructure
	node.Infra, err = infra.Run(
		ctx,
		node.Identity(),
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

	// Run the peer manager
	node.Peers, err = peers.Run(ctx, node.identity, node.Infra, node.Contacts, &node.events)
	if err != nil {
		return nil, err
	}

	// Modules
	node.runModules(ctx, modules)

	// Presence
	node.Presence = presence.Run(ctx, node.Infra, &node.events)

	err = node.Presence.Announce(ctx)
	if err != nil {
		log.Println("announce error:", err) // non-critical
	}

	// Run it
	node.process(ctx)

	return node, nil
}

func (node *Node) process(ctx context.Context) {
	go node.processQueries(ctx)
	go node.processEvents(ctx)
}

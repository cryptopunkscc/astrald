package node

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/storage"
	"time"
)

const defaultQueryTimeout = time.Minute

type Node struct {
	Config   config.Config
	identity id.Identity

	Infra    *infra.Infra
	Contacts *contacts.Manager
	Ports    *hub.Hub
	Modules  *ModuleManager

	Peers    *peers.Manager
	Store    storage.Store
	Presence *presence.Manager

	events event.Queue
}

func New(dataDir string, modules ...ModuleLoader) (*Node, error) {
	var err error
	var node = &Node{}

	// set up storage
	node.Store = storage.NewFilesystemStorage(dataDir)

	// load config
	node.Config, err = config.Load(node.Store)
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	// identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("error setting up identity: %w", err)
	}

	// hub
	node.Ports = hub.New(&node.events)

	// contacts
	node.Contacts = contacts.New(node.Store)

	// infrastructure
	node.Infra, err = infra.NewInfra(
		node.Identity(),
		node.Config.Infra,
		infra.FilteredQuerier{Querier: node, FilteredID: node.identity},
		node.Store,
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// peer manager
	node.Peers, err = peers.NewManager(node.identity, node.Infra, node.Contacts, &node.events)
	if err != nil {
		return nil, fmt.Errorf("error setting up peer manager: %w", err)
	}

	// presence
	node.Presence, err = presence.NewManager(node.Infra, &node.events)
	if err != nil {
		return nil, fmt.Errorf("error setting up presence: %w", err)
	}

	// modules
	node.Modules, err = NewModuleManager(node, modules)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	return node, nil
}

func (node *Node) Identity() id.Identity {
	return node.identity
}

func (node *Node) Alias() string {
	return node.Config.GetAlias()
}

func (node *Node) Subscribe(cancel sig.Signal) <-chan event.Event {
	return node.events.Subscribe(cancel)
}

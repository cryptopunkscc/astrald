package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

const logTag = "node"

var _ Node = &CoreNode{}

type CoreNode struct {
	config   Config
	identity id.Identity
	assets   assets.Store
	keys     assets.KeyStore

	router   *router.CoreRouter
	infra    *infra.CoreInfra
	network  *network.CoreNetwork
	tracker  *tracker.CoreTracker
	modules  *modules.CoreModules
	resolver *resolver.CoreResolver
	events   events.Queue
	routes   *router.PrefixRouter

	startedAt time.Time

	logConfig LogConfig
	logFields
}

// NewCoreNode instantiates a new node
func NewCoreNode(rootDir string) (*CoreNode, error) {
	var err error
	var node = &CoreNode{
		config: defaultConfig,
		routes: router.NewQueryRouter(),
	}

	// basic logs
	node.setupLogs()

	// assets
	node.assets, err = assets.NewFileStore(rootDir, node.log.Tag("assets"))
	if err != nil {
		return nil, err
	}

	node.keys, err = node.assets.KeyStore()
	if err != nil {
		return nil, err
	}

	// log config
	if err := node.loadLogConfig(node.assets); err != nil {
		return nil, fmt.Errorf("logger error: %w", err)
	}

	// node config
	err = node.assets.LoadYAML(configName, &node.config)
	if err != nil {
		if !errors.Is(err, assets.ErrNotFound) {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}

	// infrastructure
	node.infra, err = infra.NewCoreInfra(node, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// tracker
	node.tracker, err = tracker.NewCoreTracker(node.assets, node.infra, node.log, &node.events)
	if err != nil {
		return nil, err
	}

	// identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("error setting up identity: %w", err)
	}

	// resolver
	node.resolver = resolver.NewCoreResolver(node)

	// network
	node.network, err = network.NewCoreNetwork(node, &node.events, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up peer manager: %w", err)
	}

	// modules
	var enabled = node.config.Modules
	if enabled == nil {
		enabled = modules.RegisteredModules()
	}
	node.modules, err = modules.NewCoreModules(node, enabled, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	node.router = router.NewCoreRouter(node.log, &node.events)
	node.router.AddRoute(id.Anyone, node.Identity(), node, 100)

	node.router.SetLogRouteTrace(node.config.LogRouteTrace)

	return node, nil
}

func (node *CoreNode) Conns() *router.ConnSet {
	return node.router.Conns()
}

// Infra returns node's infrastructure component
func (node *CoreNode) Infra() infra.Infra {
	return node.infra
}

func (node *CoreNode) Tracker() tracker.Tracker {
	return node.tracker
}

func (node *CoreNode) Network() network.Network {
	return node.network
}

func (node *CoreNode) Modules() modules.Modules {
	return node.modules
}

func (node *CoreNode) Resolver() resolver.Resolver {
	return node.resolver
}

// Identity returns node's identity
func (node *CoreNode) Identity() id.Identity {
	return node.identity
}

func (node *CoreNode) Router() router.Router {
	return node.router
}

func (node *CoreNode) LocalRouter() router.LocalRouter {
	return node.routes
}

// Events returns the event queue for the node
func (node *CoreNode) Events() *events.Queue {
	return &node.events
}

func (node *CoreNode) StartedAt() time.Time {
	return node.startedAt
}

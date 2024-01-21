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
	"github.com/cryptopunkscc/astrald/resources"
	"gorm.io/gorm"
	"os"
	"time"
)

const logTag = "node"

var _ Node = &CoreNode{}

type CoreNode struct {
	identity id.Identity
	config   Config

	assets   *assets.CoreAssets
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
func NewCoreNode(nodeID id.Identity, res resources.Resources) (*CoreNode, error) {
	var err error

	if nodeID.PrivateKey() == nil {
		return nil, errors.New("private key required")
	}

	if res == nil {
		res = resources.NewMemResources()
	}

	var node = &CoreNode{
		identity: nodeID,
		config:   defaultConfig,
		routes:   router.NewPrefixRouter(true),
	}

	// basic logs
	node.setupLogs()

	node.assets = assets.NewCoreAssets(res, nil)

	// log config
	if err := node.loadLogConfig(); err != nil {
		return nil, fmt.Errorf("logger error: %w", err)
	}

	// node config
	err = node.assets.LoadYAML(configName, &node.config)
	if err != nil {
		if !errors.Is(err, resources.ErrNotFound) {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}

	// infrastructure
	node.infra, err = infra.NewCoreInfra(node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// tracker
	node.tracker, err = tracker.NewCoreTracker(node.assets, node.infra, node.log, &node.events)
	if err != nil {
		return nil, err
	}

	// check if our alias is set
	err = node.checkNodeAlias()
	if err != nil {
		node.log.Error("checkNodeAlias: %v", err)
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
	node.modules, err = modules.NewCoreModules(node, enabled, node.assets.WithPrefix("mod_"), node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	node.router = router.NewCoreRouter(node.log, &node.events)
	node.router.AddRoute(id.Anyone, node.Identity(), node, 100)

	node.router.SetLogRouteTrace(node.config.LogRouteTrace)

	return node, nil
}

func (node *CoreNode) checkNodeAlias() error {
	alias, err := node.tracker.GetAlias(node.identity)
	if (err != nil) && (!errors.Is(err, gorm.ErrRecordNotFound)) {
		return err
	}
	if alias != "" {
		return nil
	}

	alias = "localnode"

	hostname, err := os.Hostname()
	if err == nil {
		if hostname != "" && hostname != "localhost" {
			alias = hostname
		}
	}

	return node.tracker.SetAlias(node.identity, alias)
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

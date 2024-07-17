package core

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"time"
)

const logTag = "node"

var _ node.Node = &CoreNode{}

type CoreNode struct {
	identity id.Identity
	config   Config

	assets   *assets.CoreAssets
	router   *CoreRouter
	modules  *CoreModules
	resolver *CoreResolver
	auth     *CoreAuthorizer
	events   events.Queue
	routes   *routers.PrefixRouter

	startedAt time.Time

	logConfig LogConfig
	logFields

	newRouter *routers.PriorityRouter
}

func (node *CoreNode) Router() net.Router {
	return node.newRouter
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
		routes:   routers.NewPrefixRouter(true),
	}
	node.routes.EnableParams = true

	// basic logs
	node.setupLogs()

	node.assets, err = assets.NewCoreAssets(res, nil)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

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

	// authorizer
	node.auth, err = NewCoreAuthorizer(node.log.Tag("auth"))
	if err != nil {
		return nil, fmt.Errorf("error setting up authorizer: %w", err)
	}

	// resolver
	node.resolver = NewCoreResolver(node)

	// modules
	var enabled = node.config.Modules
	if enabled == nil {
		enabled = RegisteredModules()
	}
	node.modules, err = NewCoreModules(node, enabled, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	node.router = NewCoreRouter(node.log, &node.events)
	node.router.AddRoute(id.Anyone, node.Identity(), node, 100)

	node.router.SetLogRouteTrace(node.config.LogRouteTrace)

	node.newRouter = routers.NewPriorityRouter()
	node.newRouter.Add(node.router, 10)

	return node, nil
}

func (node *CoreNode) Conns() *ConnSet {
	return node.router.Conns()
}

func (node *CoreNode) Auth() node.AuthEngine {
	return node.auth
}

func (node *CoreNode) Modules() node.ModuleEngine {
	return node.modules
}

func (node *CoreNode) Resolver() node.ResolverEngine {
	return node.resolver
}

// Identity returns node's identity
func (node *CoreNode) Identity() id.Identity {
	return node.identity
}

func (node *CoreNode) LocalRouter() routers.LocalRouter {
	return node.routes
}

// Events returns the event queue for the node
func (node *CoreNode) Events() *events.Queue {
	return &node.events
}

func (node *CoreNode) StartedAt() time.Time {
	return node.startedAt
}

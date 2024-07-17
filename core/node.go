package core

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"time"
)

const logTag = "node"

var _ node.Node = &Node{}

type Node struct {
	identity id.Identity
	config   Config

	assets   *assets.CoreAssets
	router   *Router
	modules  *Modules
	resolver *Resolver
	auth     *Authorizer
	events   events.Queue

	startedAt time.Time

	logConfig LogConfig
	logFields
}

// NewNode instantiates a new node
func NewNode(nodeID id.Identity, res resources.Resources) (*Node, error) {
	var err error

	if nodeID.PrivateKey() == nil {
		return nil, errors.New("private key required")
	}

	if res == nil {
		res = resources.NewMemResources()
	}

	var node = &Node{
		identity: nodeID,
		config:   defaultConfig,
	}

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
	node.auth, err = NewAuthorizer(node.log.Tag("auth"))
	if err != nil {
		return nil, fmt.Errorf("error setting up authorizer: %w", err)
	}

	// resolver
	node.resolver = NewResolver(node)

	// modules
	var enabled = node.config.Modules
	if enabled == nil {
		enabled = RegisteredModules()
	}
	node.modules, err = NewModules(node, enabled, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	// router
	node.router = NewRouter(node.log, &node.events)
	node.router.SetLogRouteTrace(node.config.LogRouteTrace)

	return node, nil
}

func (node *Node) Router() net.Router {
	return node.router
}

func (node *Node) Auth() node.AuthEngine {
	return node.auth
}

func (node *Node) Modules() node.ModuleEngine {
	return node.modules
}

func (node *Node) Resolver() node.ResolverEngine {
	return node.resolver
}

// Identity returns node's identity
func (node *Node) Identity() id.Identity {
	return node.identity
}

// Events returns the event queue for the node
func (node *Node) Events() *events.Queue {
	return &node.events
}

func (node *Node) StartedAt() time.Time {
	return node.startedAt
}

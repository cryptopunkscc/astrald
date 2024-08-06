package core

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/resources"
	"time"
)

const logTag = "node"

var _ astral.Node = &Node{}

type Node struct {
	*Router
	identity *astral.Identity
	config   Config

	assets  *assets.CoreAssets
	modules *Modules
	events  events.Queue

	startedAt time.Time

	logConfig LogConfig
	logFields
}

// NewNode instantiates a new node
func NewNode(nodeID *astral.Identity, res resources.Resources) (*Node, error) {
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

	// router
	node.Router = NewRouter(node)

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

	// modules
	var enabled = node.config.Modules
	if enabled == nil {
		enabled = modules.Keys()
	}
	node.modules, err = NewModules(node, enabled, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	return node, nil
}

func (node *Node) Modules() *Modules {
	return node.modules
}

// Identity returns node's identity
func (node *Node) Identity() *astral.Identity {
	return node.identity
}

// Events returns the event queue for the node
func (node *Node) Events() *events.Queue {
	return &node.events
}

func (node *Node) StartedAt() time.Time {
	return node.startedAt
}

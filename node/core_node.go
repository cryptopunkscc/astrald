package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

const defaultQueryTimeout = time.Minute
const logTag = "node"

var _ Node = &CoreNode{}

type CoreNode struct {
	events     events.Queue
	assets     assets.Store
	config     Config
	logConfig  LogConfig
	identity   id.Identity
	queryQueue chan *link.Query

	log      *log.Logger
	infra    *infra.CoreInfra
	network  *network.CoreNetwork
	tracker  *tracker.CoreTracker
	services *services.CoreService
	modules  *modules.CoreModules
	resolver *resolver.CoreResolver

	rootDir string
}

// NewCoreNode instantiates a new node
func NewCoreNode(rootDir string) (*CoreNode, error) {
	var err error
	var node = &CoreNode{
		log:        log.Tag(logTag),
		rootDir:    rootDir,
		config:     defaultConfig,
		queryQueue: make(chan *link.Query),
	}

	node.assets, _ = assets.NewFileStore(rootDir, node.log.Tag("assets"))

	// logger
	if err := node.setupLogging(node.assets); err != nil {
		return nil, fmt.Errorf("logger error: %w", err)
	}

	// config
	err = node.assets.LoadYAML("node", &node.config)
	if err != nil {
		if !errors.Is(err, assets.ErrNotFound) {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}

	// identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("error setting up identity: %w", err)
	}

	// resolver
	node.resolver = resolver.NewCoreResolver(node)

	// hub
	node.services = services.NewCoreServices(&node.events, node.log)

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

	// resolve identities in logs
	log.SetFormatter(id.Identity{}, func(v interface{}) string {
		identity := v.(id.Identity)

		if node.identity.IsEqual(identity) {
			return node.log.Green() + node.Alias() + node.log.Reset()
		}

		if c := node.Resolver().DisplayName(identity); len(c) > 0 {
			return node.log.Cyan() + c + node.log.Reset()
		}

		return node.log.Cyan() + identity.Fingerprint() + node.log.Reset()
	})

	// network
	node.network, err = network.NewCoreNetwork(node.identity, node.infra, node.tracker, &node.events, node.onQuery, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up peer manager: %w", err)
	}

	// modules
	node.modules, err = modules.NewCoreModules(node, node.config.Modules, node.assets, node.log)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	return node, nil
}

// Infra returns node's infrastructure component
func (node *CoreNode) Infra() infra.Infra {
	return node.infra
}

func (node *CoreNode) Services() services.Services {
	return node.services
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

// Events returns the event queue for the node
func (node *CoreNode) Events() *events.Queue {
	return &node.events
}

// Alias returns node's alias
func (node *CoreNode) Alias() string {
	return node.config.Alias
}

// SetAlias sets the node alias
func (node *CoreNode) SetAlias(alias string) error {
	node.config.Alias = alias
	return node.assets.StoreYAML(configName, node.config)
}

func (node *CoreNode) setupLogging(assets assets.Store) error {
	if err := assets.LoadYAML("log", &node.logConfig); err != nil {
		return nil
	}

	for tag, level := range node.logConfig.TagLevels {
		log.SetTagLevel(tag, level)
	}
	for tag, color := range node.logConfig.TagColors {
		log.SetTagColor(tag, color)
	}
	log.HideDate = node.logConfig.HideDate
	log.Level = node.logConfig.Level

	return nil
}

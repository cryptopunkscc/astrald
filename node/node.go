package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"os"
	"path"
	"path/filepath"
	"time"
)

const defaultQueryTimeout = time.Minute
const dbFileName = "astrald.db"

type Node struct {
	events      event.Queue
	configStore config.Store
	config      *Config
	logConfig   *LogConfig
	database    *db.Database
	identity    id.Identity
	queryQueue  chan *link.Query

	Infra    *infra.Infra
	Tracker  *tracker.Tracker
	Contacts *contacts.Manager
	Ports    *hub.Hub
	Modules  *ModuleManager
	Peers    *peers.Manager
	Presence *presence.Manager

	rootDir string
}

var log = _log.Tag("node")

// New instantiates a new node
func New(rootDir string, modules ...ModuleLoader) (*Node, error) {
	var err error
	var node = &Node{
		rootDir:    rootDir,
		queryQueue: make(chan *link.Query),
	}

	node.configStore, _ = config.NewFileStore(path.Join(rootDir, "config"))

	// setup logger
	if err := node.setupLogging(node.configStore); err != nil {
		return nil, fmt.Errorf("logger error: %w", err)
	}

	// load config
	node.config, err = LoadConfig(node.configStore)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}

	// setup database
	var dbInit bool
	dbFile := filepath.Join(rootDir, dbFileName)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Log("creating database at %s", dbFile)
		dbInit = true
	}

	node.database, err = db.NewFileDatabase(dbFile)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}

	if dbInit {
		if err := tracker.InitDatabase(node.database); err != nil {
			return nil, fmt.Errorf("tracker: %w", err)
		}
		if err := contacts.InitDatabase(node.database); err != nil {
			return nil, fmt.Errorf("contacts: %w", err)
		}

		if err := os.Chmod(dbFile, 0600); err != nil {
			log.Error("cannot set 0600 mode on the database file: %s", err)
		}
	}

	// identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("error setting up identity: %w", err)
	}

	// hub
	node.Ports = hub.New(&node.events)

	// infrastructure
	node.Infra, err = infra.New(
		node.Identity(),
		node.configStore,
		infra.FilteredQuerier{Querier: node, FilteredID: node.identity},
		node.RootDir(),
	)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// tracker
	node.Tracker, err = tracker.New(node.database, node.Infra)
	if err != nil {
		return nil, err
	}

	// contacts
	node.Contacts, err = contacts.New(node.database, &node.events)
	if err != nil {
		return nil, err
	}

	_log.SetFormatter(id.Identity{}, func(v interface{}) string {
		identity := v.(id.Identity)
		if c, err := node.Contacts.Find(identity); err == nil {
			if c.Alias() != "" {
				return log.Cyan() + c.Alias() + log.Reset()
			}
		}

		return log.Green() + identity.Fingerprint() + log.Reset()
	})

	// peer manager
	node.Peers, err = peers.NewManager(node.identity, node.Infra, node.Tracker, &node.events, node.onQuery)
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

// RootDir returns node's root directory where all node-related files are stored
func (node *Node) RootDir() string {
	return node.rootDir
}

// Identity returns node's identity
func (node *Node) Identity() id.Identity {
	return node.identity
}

// Events returns the event queue for the node
func (node *Node) Events() *event.Queue {
	return &node.events
}

// Alias returns node's alias
func (node *Node) Alias() string {
	return node.config.Alias()
}

// SetAlias sets the node alias
func (node *Node) SetAlias(alias string) error {
	return node.config.SetAlias(alias)
}

// ConfigStore returns config storage for the node
func (node *Node) ConfigStore() config.Store {
	return node.configStore
}

func (node *Node) setupLogging(store config.Store) error {
	node.logConfig = &LogConfig{}

	if err := store.LoadYAML("log", node.logConfig); err != nil {
		return nil
	}

	for tag, level := range node.logConfig.TagLevels {
		_log.SetTagLevel(tag, level)
	}
	for tag, color := range node.logConfig.TagColors {
		_log.SetTagColor(tag, color)
	}
	_log.HideDate = node.logConfig.HideDate
	_log.Level = node.logConfig.Level

	return nil
}

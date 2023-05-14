package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"os"
	"path"
	"path/filepath"
	"time"
)

const defaultQueryTimeout = time.Minute
const dbFileName = "astrald.db"

type CoreNode struct {
	events      event.Queue
	configStore config.Store
	config      *Config
	logConfig   *LogConfig
	database    *db.Database
	identity    id.Identity
	queryQueue  chan *link.Query

	infra    *infra.CoreInfra
	network  *network.Network
	tracker  *tracker.CoreTracker
	contacts *contacts.Manager
	services *services.Manager
	modules  *ModuleManager

	rootDir string
}

func (node *CoreNode) Modules() *ModuleManager {
	return node.modules
}

// New instantiates a new node
func New(rootDir string, modules ...ModuleLoader) (*CoreNode, error) {
	var err error
	var node = &CoreNode{
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
	node.services = services.New(&node.events)

	// infrastructure
	node.infra, err = infra.NewCoreInfra(node, node.configStore)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// tracker
	node.tracker, err = tracker.New(node.database, node.infra)
	if err != nil {
		return nil, err
	}

	// contacts
	node.contacts, err = contacts.New(node.database, &node.events)
	if err != nil {
		return nil, err
	}

	// format identities in logs
	_log.SetFormatter(id.Identity{}, func(v interface{}) string {
		identity := v.(id.Identity)

		if node.identity.IsEqual(identity) {
			return log.Green() + node.Alias() + log.Reset()
		}

		if c, err := node.contacts.Find(identity); err == nil {
			if c.Alias() != "" {
				return log.Cyan() + c.Alias() + log.Reset()
			}
		}

		return log.Green() + identity.Fingerprint() + log.Reset()
	})

	// peer manager
	node.network, err = network.NewNetwork(node.identity, node.infra, node.tracker, &node.events, node.onQuery)
	if err != nil {
		return nil, fmt.Errorf("error setting up peer manager: %w", err)
	}

	// modules
	node.modules, err = NewModuleManager(node, modules)
	if err != nil {
		return nil, fmt.Errorf("error creating module manager: %w", err)
	}

	return node, nil
}

// Infra returns node's infrastructure component
func (node *CoreNode) Infra() infra.Infra {
	return node.infra
}

func (node *CoreNode) Services() Services {
	return node.services
}

func (node *CoreNode) Contacts() Contacts {
	return node.contacts
}

func (node *CoreNode) Tracker() Tracker {
	return node.tracker
}

func (node *CoreNode) Network() Network {
	return node.network
}

// RootDir returns node's root directory where all node-related files are stored
func (node *CoreNode) RootDir() string {
	return node.rootDir
}

// Identity returns node's identity
func (node *CoreNode) Identity() id.Identity {
	return node.identity
}

// Events returns the event queue for the node
func (node *CoreNode) Events() *event.Queue {
	return &node.events
}

// Alias returns node's alias
func (node *CoreNode) Alias() string {
	return node.config.Alias()
}

// SetAlias sets the node alias
func (node *CoreNode) SetAlias(alias string) error {
	return node.config.SetAlias(alias)
}

// ConfigStore returns config storage for the node
func (node *CoreNode) ConfigStore() config.Store {
	return node.configStore
}

func (node *CoreNode) setupLogging(store config.Store) error {
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

	if fileStore, ok := node.configStore.(*config.FileStore); ok {
		fileStore.Errorv = log.Tag("config").Errorv
	}

	return nil
}

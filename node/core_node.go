package node

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/modules"
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
const logTag = "node"

var _ Node = &CoreNode{}

type CoreNode struct {
	events      event.Queue
	configStore config.Store
	config      Config
	logConfig   LogConfig
	database    *db.Database
	identity    id.Identity
	queryQueue  chan *link.Query

	log      *log.Logger
	infra    *infra.CoreInfra
	network  *network.CoreNetwork
	tracker  *tracker.CoreTracker
	contacts *contacts.CoreContacts
	services *services.CoreService
	modules  *modules.CoreModules

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

	node.configStore, _ = config.NewFileStore(path.Join(rootDir, "config"))

	// setup logger
	if err := node.setupLogging(node.configStore); err != nil {
		return nil, fmt.Errorf("logger error: %w", err)
	}

	// load config
	err = node.configStore.LoadYAML("node", &node.config)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error loading config: %w", err)
		}
	}

	// setup database
	var dbInit bool
	dbFile := filepath.Join(rootDir, dbFileName)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		node.log.Log("creating database at %s", dbFile)
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
			node.log.Error("cannot set 0600 mode on the database file: %s", err)
		}
	}

	// identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("error setting up identity: %w", err)
	}

	// hub
	node.services = services.NewCoreServices(&node.events, node.log)

	// infrastructure
	node.infra, err = infra.NewCoreInfra(node, node.configStore, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up infrastructure: %w", err)
	}

	// tracker
	node.tracker, err = tracker.NewCoreTracker(node.database, node.infra, node.log)
	if err != nil {
		return nil, err
	}

	// contacts
	node.contacts, err = contacts.NewCoreContacts(node.database, &node.events)
	if err != nil {
		return nil, err
	}

	// format identities in logs
	log.SetFormatter(id.Identity{}, func(v interface{}) string {
		identity := v.(id.Identity)

		if node.identity.IsEqual(identity) {
			return node.log.Green() + node.Alias() + node.log.Reset()
		}

		if c, err := node.contacts.Find(identity); err == nil {
			if c.Alias() != "" {
				return node.log.Cyan() + c.Alias() + node.log.Reset()
			}
		}

		return node.log.Green() + identity.Fingerprint() + node.log.Reset()
	})

	// peer manager
	node.network, err = network.NewCoreNetwork(node.identity, node.infra, node.tracker, &node.events, node.onQuery, node.log)
	if err != nil {
		return nil, fmt.Errorf("error setting up peer manager: %w", err)
	}

	// modules
	node.modules, err = modules.NewCoreModules(node, node.config.Modules, node.configStore, node.log)
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

func (node *CoreNode) Contacts() contacts.Contacts {
	return node.contacts
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
	return node.config.Alias
}

// SetAlias sets the node alias
func (node *CoreNode) SetAlias(alias string) error {
	node.config.Alias = alias
	return node.configStore.StoreYAML(configName, node.config)
}

func (node *CoreNode) setupLogging(store config.Store) error {
	if err := store.LoadYAML("log", &node.logConfig); err != nil {
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

	if fileStore, ok := node.configStore.(*config.FileStore); ok {
		fileStore.Errorv = node.log.Tag("config").Errorv
	}

	return nil
}

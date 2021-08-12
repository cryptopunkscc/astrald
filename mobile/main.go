package astralandroid

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/java"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_ "github.com/cryptopunkscc/astrald/services/appsupport/tcp"
	"github.com/cryptopunkscc/astrald/services/fs"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Exit statuses
const (
	ExitSuccess     = int(iota) // Normal exit
	ExitHelp                    // Help was invoked
	ExitNodeError               // Node reported an error
	ExitForced                  // User forced shutdown with double SIGINT
	ExitConfigError             // An invalid or non-existent config file provided
)

const defaultConfigFilename = "astrald.conf"
const defaultIdentityFilename = "id"
const defaultPort = 1791

type configType struct {
	Port int
}

var config configType
var defaultConfig = configType{
	Port: defaultPort,
}

var shutdown context.CancelFunc

func loadID(astralDir string) *id.ECIdentity {
	idPath := filepath.Join(astralDir, defaultIdentityFilename)

	// Try to load an existing identity
	idBytes, err := ioutil.ReadFile(idPath)
	if err == nil {
		id, err := id.ECIdentityFromBytes(idBytes)
		if err != nil {
			panic(err)
		}
		return id
	}

	// The only acceptable error is ErrNotExist
	if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	// Generate a new identity
	log.Println("generating new node identity...")
	id, err := id.GenerateECIdentity()
	if err != nil {
		panic(err)
	}

	// Save the new identity
	_ = ioutil.WriteFile(idPath, id.PrivateKey().Serialize(), 0600)

	return id
}

func Start(astralHome string) {
	// Figure out the config path
	log.Println("log Staring astrald")
	fs.AstralHome = astralHome

	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	} else {
		configPath = filepath.Join(astralHome, defaultConfigFilename)
	}

	// Load the config file
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		config = defaultConfig
		configBytes, _ = yaml.Marshal(&config)
		_ = ioutil.WriteFile(configPath, configBytes, 0600)
	} else {
		// Parse config file
		err = yaml.Unmarshal(configBytes, &config)
		if err != nil {
			log.Println("error parsing config file:", err)
			os.Exit(ExitConfigError)
		}
	}

	// Set up app execution context
	var ctx context.Context
	ctx, shutdown = context.WithCancel(context.Background())
	//ctx, _ := context.WithCancel(context.Background())

	// Instantiate the node
	node := node.New(
		loadID(astralHome),
		config.Port,
	)

	// Run the node
	err = node.Run(ctx)

	time.Sleep(50 * time.Millisecond)

	// Check results
	if err != nil {
		log.Printf("error: %s\n", err)
		os.Exit(ExitNodeError)
	} else {
		log.Printf("success.\n")
		os.Exit(ExitSuccess)
	}
}

func Stop() {
	shutdown()
}

func Register(
	name string,
	srv astraljava.Service,
) error {
	return node.RegisterService(
		name,
		serviceAdapter{delegate: srv},
	)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_ "github.com/cryptopunkscc/astrald/services/appsupport"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

func astralDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(ExitConfigError)
	}

	dir := filepath.Join(cfgDir, "astrald")
	os.MkdirAll(dir, 0700)

	return dir
}

func loadID() *id.ECIdentity {
	idPath := filepath.Join(astralDir(), defaultIdentityFilename)

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

func main() {
	// Figure out the config path
	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	} else {
		configPath = filepath.Join(astralDir(), defaultConfigFilename)
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
			fmt.Println("error parsing config file:", err)
			os.Exit(ExitConfigError)
		}
	}

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// Instantiate the node
	node := node.New(
		loadID(),
		config.Port,
	)

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			log.Println("shutting down...")
			shutdown()

			<-sigCh
			log.Println("forcing shutdown...")
			os.Exit(ExitForced)
		}
	}()

	// Run the node
	err = node.Run(ctx)

	time.Sleep(50 * time.Millisecond)

	// Check results
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(ExitNodeError)
	} else {
		fmt.Printf("success.\n")
		os.Exit(ExitSuccess)
	}
}

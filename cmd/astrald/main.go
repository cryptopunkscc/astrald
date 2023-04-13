package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/cryptopunkscc/astrald/infra/tor/system"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/agent"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/keepalive"
	"github.com/cryptopunkscc/astrald/mod/optimizer"
	"github.com/cryptopunkscc/astrald/mod/reflectlink"
	"github.com/cryptopunkscc/astrald/mod/roam"
	"github.com/cryptopunkscc/astrald/mod/tcpfwd"
	"github.com/cryptopunkscc/astrald/node"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

// Exit statuses
const (
	ExitSuccess     = int(iota) // Normal exit
	ExitHelp                    // Help was invoked
	ExitNodeError               // Node reported an error
	ExitForced                  // User forced shutdown with double SIGINT
	ExitConfigError             // An invalid or non-existent config file provided
)

func astralDir() string {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("error fetching user's config dir:", err)
		os.Exit(ExitConfigError)
	}

	dir := filepath.Join(cfgDir, "astrald")

	return dir
}

func main() {
	var err error
	var astralRoot = astralDir()

	flag.StringVar(&astralRoot, "datadir", astralRoot, "set data directory")
	flag.Parse()

	if strings.HasPrefix(astralRoot, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			astralRoot = filepath.Join(homeDir, astralRoot[2:])
		}
	}

	// make sure root directory exists
	os.MkdirAll(astralRoot, 0700)

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			log.Log("shutting down...")
			shutdown()

			<-sigCh
			log.Log("forcing shutdown...")
			os.Exit(ExitForced)
		}
	}()

	// start the node
	node, err := node.New(
		astralRoot,
		admin.Loader{},
		apphost.Loader{},
		connect.Loader{},
		gateway.Loader{},
		reflectlink.Loader{},
		roam.Loader{},
		optimizer.Loader{},
		keepalive.Loader{},
		agent.Loader{},
		tcpfwd.Loader{},
	)
	if err != nil {
		fmt.Println("init error:", err)
		os.Exit(ExitNodeError)
	}

	// Run the node
	if err := node.Run(ctx); err != nil {
		fmt.Printf("run error: %s\n", err)
		os.Exit(ExitNodeError)
	}

	os.Exit(ExitSuccess)
}

package main

import (
	"context"
	"fmt"
	_ "github.com/cryptopunkscc/astrald/infra/tor/system"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/linkinfo"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/optimizer"
	"github.com/cryptopunkscc/astrald/mod/roam"
	"github.com/cryptopunkscc/astrald/node"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	os.MkdirAll(dir, 0700)

	return dir
}

func main() {
	var err error
	var astralRoot = astralDir()

	// Figure out the config path
	if len(os.Args) > 1 {
		astralRoot = os.Args[1]
	}

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

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

	// start the node
	node, err := node.New(
		astralRoot,
		admin.Loader{},
		apphost.Loader{},
		connect.Loader{},
		gateway.Loader{},
		linkinfo.Loader{},
		roam.Loader{},
		contacts.Loader{},
		optimizer.Loader{},
		nat.Loader{},
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

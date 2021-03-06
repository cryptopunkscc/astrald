package main

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/cryptopunkscc/astrald/infra/tor/system"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/connect"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/info"
	"github.com/cryptopunkscc/astrald/mod/linkback"
	"github.com/cryptopunkscc/astrald/mod/roam"
	"github.com/cryptopunkscc/astrald/node"
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
	astralRoot := astralDir()

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
	_, err := node.Run(
		ctx,
		astralRoot,
		admin.Admin{},
		&apphost.Module{},
		connect.Connect{},
		gateway.Gateway{},
		info.Info{},
		linkback.LinkBack{},
		&roam.Roam{},
	)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()

	time.Sleep(50 * time.Millisecond)

	// Check results
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			fmt.Printf("error: %s\n", err)
			os.Exit(ExitNodeError)
		}
	}

	fmt.Printf("success.\n")
	os.Exit(ExitSuccess)
}

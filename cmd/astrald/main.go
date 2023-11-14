package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/node"
	"os"
	"os/signal"
	"path/filepath"
	_debug "runtime/debug"
	"strings"
	"syscall"
	"time"
)

var astralRoot string
var version bool

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
	astralRoot = astralDir()

	flag.StringVar(&astralRoot, "datadir", astralRoot, "set data directory")
	flag.BoolVar(&version, "v", false, "show version")
	flag.Parse()

	if version {
		if info, ok := _debug.ReadBuildInfo(); ok {
			for _, setting := range info.Settings {
				if setting.Key == "vcs.revision" {
					fmt.Println("commit", setting.Value)
					os.Exit(0)
				}
			}
		}
		fmt.Println("no version info available")
		os.Exit(0)
	}

	if strings.HasPrefix(astralRoot, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			astralRoot = filepath.Join(homeDir, astralRoot[2:])
		}
	}

	// make sure root directory exists
	os.MkdirAll(astralRoot, 0700)

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	debug.LogDir = astralRoot
	defer debug.SaveLog(func(p any) {
		debug.SigInt(p)
		time.Sleep(time.Second) // give components time to exit cleanly
	})

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			fmt.Println("shutting down...")
			shutdown()

			<-sigCh
			fmt.Println("forcing shutdown...")
			os.Exit(ExitForced)
		}
	}()

	// start the node
	node, err := node.NewCoreNode(astralRoot)
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

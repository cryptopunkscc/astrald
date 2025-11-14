package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/cryptopunkscc/astrald/mod/all"
)

// Exit statuses
const (
	ExitSuccess   = int(iota) // Normal exit
	ExitNodeError             // Node reported an error
	ExitForced                // User forced shutdown with double SIGINT
)

func main() {
	var args = parseArgs()

	if args.Version {
		os.Exit(printVersion())
	}

	// set up node execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		<-sigCh
		fmt.Println("shutting down...")
		shutdown()

		<-sigCh
		fmt.Println("forcing shutdown...")
		os.Exit(ExitForced)
	}()

	// run the node
	if err := run(ctx, args); err != nil {
		fmt.Printf("node error: %s\n", err)
		os.Exit(ExitNodeError)
	}

	os.Exit(ExitSuccess)
}

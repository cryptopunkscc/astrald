package main

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/warpdrived"
	"github.com/cryptopunkscc/astrald/lib/wrapper/apphost"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// The warpdrive launcher for desktop.
func main() {
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
			os.Exit(0)
		}
	}()

	err := warpdrived.Desktop().Run(ctx, apphost.Adapter{})

	code := 0
	if err != nil {
		log.Print(fmt.Println("cannot run server", err))
		code = 1
		shutdown()
	}

	time.Sleep(50 * time.Millisecond)

	os.Exit(code)
}

package main

import (
	"context"
	warpdrive "github.com/cryptopunkscc/astrald/app/warpdrive/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

	warpdrive.RunService(ctx)

	time.Sleep(50 * time.Millisecond)

	os.Exit(0)
}

package main

import (
	"context"
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/server"
	"github.com/cryptopunkscc/astrald/lib/wrapper/apphost"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
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

	srv := &server.Server{}
	err := srv.Run(ctx, apphost.Adapter{})

	code := 0
	if err != nil {
		err = warpdrive.Error(err, "cannot run server")
		log.Println(err)
		code = 1
		shutdown()
	}

	time.Sleep(50 * time.Millisecond)

	os.Exit(code)
}

package main

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/wrapper/apphost"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	var err error

	// Set up app execution context
	ctx, shutdown := context.WithCancel(context.Background())

	// init dispatcher
	d := warpdrive.Dispatcher{
		Context:    ctx,
		LogPrefix:  "[CLI]",
		Api:        apphost.Adapter{},
		Authorized: true,
	}

	// resolve identity
	identity, err := d.Api.Resolve("localnode")
	if err != nil {
		log.Panicln(warpdrive.Error(err, "cannot resolve local node id"))
		return
	}
	d.CallerId = identity.String()

	// setup connection
	pr, pw := io.Pipe()
	rw := &stdReadWrite{pr, os.Stdout}
	d.Conn = rw
	d.Endec = cslq.NewEndec(rw)

	// run cli
	go func() {
		err = warpdrive.Cli(&d)
		if err != nil {
			d.Panicln(err)
		} else {
			os.Exit(0)
		}
	}()

	// handler args
	switch len(os.Args) > 1 {
	case true:
		// format application arguments and pass to cli
		args := strings.Join(os.Args[1:], " ")
		//go func() {
		_, err := fmt.Fprint(pw, "prompt-off", "\n", args, "\n", "exit", "\n")
		if err != nil {
			d.Println(warpdrive.Error(err, "cannot write args"))
		}
		//}()
	case false:
		// switch to interactive mode, pass std in to cli
		go func() {
			_, err := io.Copy(pw, os.Stdin)
			if err != nil {
				d.Println(warpdrive.Error(err, "cannot copy std in"))
			}
		}()
	}

	// Trap ctrl+c
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT)
	go func() {
		for {
			<-sigCh
			println()
			log.Println("shutting down...")
			shutdown()

			<-sigCh
			println()
			log.Println("forcing shutdown...")
			os.Exit(0)
		}
	}()

	code := 0
	if err != nil {
		log.Println("cannot run server", err)
		code = 1
		shutdown()
	}

	<-ctx.Done()

	pw.Close()

	time.Sleep(50 * time.Millisecond)

	os.Exit(code)
}

type stdReadWrite struct {
	io.Reader
	io.Writer
}

func (s stdReadWrite) Close() error {
	return nil
}

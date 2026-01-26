package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
	apphost "github.com/cryptopunkscc/astrald/mod/apphost/client"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
)

func main() {
	var accept string
	flag.StringVar(&accept, "a", "", "accept query")
	flag.Parse()

	var ctx = astrald.NewContext()

	server, err := astrald.Listen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(1)
	}

	err = apphost.RegisterHandler(ctx, server.Endpoint(), server.AuthToken())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for {
		query, err := server.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		caller, _ := dircli.GetAlias(ctx, query.Caller())
		if caller == "" {
			caller = query.Caller().String()
		}

		if query.Query() == accept {
			fmt.Fprintf(os.Stderr, "accepted [%s] %s from %s\n", query.Nonce(), query.Query(), caller)
			conn := query.Accept()

			go func() {
				defer conn.Close()
				io.Copy(conn, os.Stdin)
			}()

			io.Copy(os.Stdout, conn)
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "ignored [%s] %s from %s\n", query.Nonce(), query.Query(), caller)

		query.Skip()
	}
}

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func main() {
	var accept string
	flag.StringVar(&accept, "a", "", "accept query")
	flag.Parse()

	var ctx = astrald.NewContext()

	l, err := astrald.AppHost().RegisterHandler(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for {
		query, err := l.Next()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		caller, _ := astrald.Dir().GetAlias(ctx, query.Caller())
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

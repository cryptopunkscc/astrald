package main

import (
	"flag"
	"fmt"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"io"
	"os"
)

func main() {
	var accept string
	flag.StringVar(&accept, "a", "", "accept query")
	flag.Parse()

	client, err := apphost.NewDefaultClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	l, err := client.Listen()
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

		if query.Query() == accept {
			fmt.Fprintf(os.Stderr, "%s:%s (accepted)\n", query.Caller(), query.Query())
			conn, err := query.Accept()
			if err != nil {
				panic(err)
			}

			go func() {
				defer conn.Close()
				io.Copy(conn, os.Stdin)
			}()

			io.Copy(os.Stdout, conn)
			os.Exit(0)
		}

		fmt.Fprintf(os.Stderr, "%s:%s (ignored)\n", query.Caller(), query.Query())

		query.Close()
	}
}

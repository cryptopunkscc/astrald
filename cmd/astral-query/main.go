package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %s <query>\n", os.Args[0])
		return
	}

	client, err := apphost.NewDefaultClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var caller, target *astral.Identity

	c, t, query := arl.Split(os.Args[1])
	if len(c) > 0 {
		caller, err = client.ResolveIdentity(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		caller = client.GuestID
	}

	if len(t) > 0 {
		target, err = client.ResolveIdentity(t)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		target = client.GuestID
	}

	s, err := client.Session()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	conn, err := s.Query(caller, target, query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	go func() {
		defer conn.Close()
		io.Copy(conn, os.Stdin)
	}()

	io.Copy(os.Stdout, conn)
}

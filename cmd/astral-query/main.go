package main

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
	"os"
	"strings"
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

	c, t, q := arl.Split(os.Args[1])
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

	var args = os.Args[2:]
	var params = map[string]string{}
	for len(args) >= 2 {
		key := args[0]
		val := args[1]
		if !strings.HasPrefix(key, "-") || len(key) < 2 {
			fmt.Fprintf(os.Stderr, "error: unexpected argument %s\n", key)
			os.Exit(2)
		}
		params[key[1:]] = val
		args = args[2:]
	}

	if len(params) > 0 {
		s, err := query.Marshal(params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		q = q + "?" + s
	}

	s, err := client.Session()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	conn, err := s.Query(caller, target, q)
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

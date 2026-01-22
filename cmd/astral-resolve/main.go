package main

import (
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: astral-resolve <name>")
		return
	}

	var ctx = astrald.NewContext()

	identity, err := dircli.ResolveIdentity(ctx, os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	fmt.Println(identity)
}

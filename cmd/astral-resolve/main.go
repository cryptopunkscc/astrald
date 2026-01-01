package main

import (
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: astral-resolve <name>")
		return
	}

	identity, err := astrald.Dir().ResolveIdentity(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	fmt.Println(identity)
}

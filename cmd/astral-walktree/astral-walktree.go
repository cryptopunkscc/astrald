package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

func walk(ctx *astral.Context, node tree.Node, depth int) error {
	fmt.Print(strings.Repeat("  ", depth))

	if node.Name() == "" {
		fmt.Println("(root)")
	} else {
		fmt.Println("/" + node.Name())
	}

	sub, err := node.Sub(ctx)
	if err != nil {
		return err
	}

	for name := range sub {
		err = walk(ctx, sub[name], depth+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ctx := astrald.NewContext()

	var tree *astrald.TreeClient

	if len(os.Args) <= 1 {
		tree = astrald.Tree()
	} else {
		targetID, err := astrald.Dir().ResolveIdentity(ctx, os.Args[1])
		if err != nil {
			fatal("resolve identity: %v\n", err)
		}

		tree = astrald.NewTreeClient(astrald.DefaultClient(), targetID)
	}

	err := walk(ctx, tree.Root(), 0)
	if err != nil {
		fatal("walk error: %v\n", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

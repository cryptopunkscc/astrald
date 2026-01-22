package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
	"github.com/cryptopunkscc/astrald/mod/tree"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
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
	client := treecli.Default()

	// parse the args
	if len(os.Args) > 1 {
		targetID, err := dircli.ResolveIdentity(ctx, os.Args[1])
		if err != nil {
			fatal("resolve identity: %v\n", err)
		}
		client = treecli.New(targetID, nil)
	}

	// walk the tree
	err := walk(ctx, client.Root(), 0)
	if err != nil {
		fatal("walk error: %v\n", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

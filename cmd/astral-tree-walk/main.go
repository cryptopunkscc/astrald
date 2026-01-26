package main

import (
	"encoding"
	"fmt"
	"os"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	dircli "github.com/cryptopunkscc/astrald/mod/dir/client"
	"github.com/cryptopunkscc/astrald/mod/tree"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
)

func walk(ctx *astral.Context, node tree.Node, path []string) error {
	fmt.Print(strings.Repeat("  ", len(path)))

	if len(path) > 0 {
		fmt.Print(path[len(path)-1])
	}

	object, err := tree.Get[astral.Object](ctx, node)
	if err == nil {
		var text []byte
		t := object.ObjectType()

		if m, ok := object.(encoding.TextMarshaler); ok {
			text, _ = m.MarshalText()
		} else {
			text = []byte("[data]")
		}

		fmt.Printf(" = \"%s\" [%s]\n", string(text), t)
	} else {
		fmt.Println()
	}

	sub, err := node.Sub(ctx)
	if err != nil {
		return err
	}

	for name := range sub {
		err = walk(ctx, sub[name], append(path, name))
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	var err error
	var targetID *astral.Identity
	var target string

	ctx := astrald.NewContext()
	client := treecli.Default()

	// parse the args
	if len(os.Args) > 1 {
		target = os.Args[1]
		targetID, err = dircli.ResolveIdentity(ctx, target)
		if err != nil {
			fatal("resolve identity: %v\n", err)
		}
		client = treecli.New(targetID, nil)
	}
	if target == "" {
		target = "localnode"
	}

	alias, err := dircli.GetAlias(ctx, targetID)
	if err == nil {
		fmt.Print(alias)
	} else {
		fmt.Print(target)
	}

	// walk the tree
	err = walk(ctx, client.Root(), nil)
	if err != nil {
		fatal("walk error: %v\n", err)
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/cryptopunkscc/astrald/lib/astrald"
)

const blockSize = 2 << 14

func main() {
	var repo string
	var alloc int
	var target string

	flag.StringVar(&repo, "repo", "", "target repository")
	flag.IntVar(&alloc, "alloc", 0, "allocate space upfront")
	flag.StringVar(&target, "target", "localnode", "target node")

	flag.Parse()

	ctx := astrald.NewContext()

	targetID, err := astrald.Dir().ResolveIdentity(ctx, target)
	if err != nil {
		fatal("resolve target: %v\n", err)
	}

	w, err := astrald.
		NewObjectsClient(targetID, nil).
		Create(ctx, repo, alloc)
	if err != nil {
		fatal("create: %v", err)
	}

	var buf = make([]byte, blockSize)

	_, err = io.CopyBuffer(w, os.Stdin, buf)
	if err != nil {
		fatal("read: %v", err)
	}

	objectID, err := w.Commit()
	if err != nil {
		fatal("commit: %v", err)
	}

	fmt.Println(objectID)
}

func fatal(f string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+f+"\n", args...)
	os.Exit(1)
}

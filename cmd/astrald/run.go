package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/resources"
)

const resNodeKey = "node_key"

func run(ctx context.Context, args *Args) error {
	nodeRes, err := setupResources(args)
	if err != nil {
		return err
	}

	nodeID, err := loadNodeIdentity(nodeRes)
	if err != nil {
		return err
	}

	// run the node
	coreNode, err := core.NewNode(nodeID, nodeRes)
	if err != nil {
		return err
	}

	return coreNode.Run(ctx)
}

func setupResources(args *Args) (resources.Resources, error) {
	if args.Ghost {
		mem := resources.NewMemResources()
		mem.Write("log.yaml", []byte("level: 2"))
		return mem, nil
	}

	nodeRes, err := resources.NewFileResources(args.NodeRoot, true)
	if err != nil {
		return nil, err
	}

	if len(args.DBRoot) > 0 {
		nodeRes.SetDatabaseRoot(args.DBRoot)
	}

	// make sure root directory exists
	err = os.MkdirAll(args.NodeRoot, 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating node directory: %s\n", err)
	}

	// set directory for saving crash logs
	debug.LogDir = args.NodeRoot
	defer debug.SaveLog(func(p any) {
		debug.SigInt(p)
		time.Sleep(time.Second) // give components time to exit cleanly
	})

	return nodeRes, err
}

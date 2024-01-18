package main

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"os"
	"time"
)

const resNodeIdentity = "node_identity"

func run(ctx context.Context, args *Args) error {
	nodeRes, err := setupResources(args)
	if err != nil {
		return err
	}

	nodeID, err := setupNodeIdentity(nodeRes)
	if err != nil {
		return err
	}

	// run the node
	coreNode, err := node.NewCoreNode(nodeID, nodeRes)
	if err != nil {
		return err
	}

	return coreNode.Run(ctx)
}

func setupResources(args *Args) (resources.Resources, error) {
	if args.Ghost {
		return resources.NewMemResources(), nil
	}

	nodeRes, err := resources.NewFileResources(args.NodeRoot, true)
	if err != nil {
		return nil, err
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

// setupNodeIdentity reads node's identity from resources or generates one if needed
func setupNodeIdentity(resources resources.Resources) (id.Identity, error) {
	key, err := resources.Read(resNodeIdentity)
	if err == nil {
		return id.ParsePrivateKey(key)
	}

	nodeID, err := id.GenerateIdentity()
	if err != nil {
		return id.Identity{}, err
	}

	err = resources.Write(resNodeIdentity, nodeID.PrivateKey().Serialize())
	if err != nil {
		return id.Identity{}, err
	}

	return nodeID, nil
}

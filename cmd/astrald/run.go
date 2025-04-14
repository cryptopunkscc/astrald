package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/keys"
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

// setupNodeIdentity reads node's identity from resources or generates one if needed
func setupNodeIdentity(resources resources.Resources) (*astral.Identity, error) {
	keyBytes, err := resources.Read(resNodeIdentity)
	if err == nil {
		if len(keyBytes) == 32 {
			return astral.IdentityFromPrivKeyBytes(keyBytes)
		}

		var pk keys.PrivateKey

		objType, payload, err := astral.OpenCanonical(bytes.NewReader(keyBytes))
		switch {
		case err != nil:
			return nil, err
		case objType != pk.ObjectType():
			return nil, fmt.Errorf("invalid object type: %s", objType)
		}

		_, err = pk.ReadFrom(payload)
		if err != nil {
			return nil, err
		}

		return astral.IdentityFromPrivKeyBytes(pk.Bytes)
	}

	nodeID, err := astral.GenerateIdentity()
	if err != nil {
		return nil, err
	}

	var buf = &bytes.Buffer{}

	pk := &keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: nodeID.PrivateKey().Serialize(),
	}

	_, err = astral.WriteCanonical(buf, pk)
	if err != nil {
		return nil, err
	}

	err = resources.Write(resNodeIdentity, buf.Bytes())
	if err != nil {
		return nil, err
	}

	return nodeID, nil
}

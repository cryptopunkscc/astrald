package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/core"
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
	coreNode, err := core.NewCoreNode(nodeID, nodeRes)
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
	keyBytes, err := resources.Read(resNodeIdentity)
	if err == nil {
		if len(keyBytes) == 32 {
			return id.ParsePrivateKey(keyBytes)
		}

		var r = bytes.NewReader(keyBytes)
		err = adc.ExpectHeader(r, keys.PrivateKeyDataType)
		if err != nil {
			return id.Identity{}, err
		}
		var pk keys.PrivateKey
		err = cslq.Decode(r, "v", &pk)
		if err != nil {
			return id.Identity{}, err
		}
		return id.ParsePrivateKey(pk.Bytes)
	}

	nodeID, err := id.GenerateIdentity()
	if err != nil {
		return id.Identity{}, err
	}

	var buf = &bytes.Buffer{}

	err = cslq.Encode(buf, "vv", adc.Header(keys.PrivateKeyDataType), keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: nodeID.PrivateKey().Serialize(),
	})
	if err != nil {
		return id.Identity{}, err
	}

	err = resources.Write(resNodeIdentity, buf.Bytes())
	if err != nil {
		return id.Identity{}, err
	}

	return nodeID, nil
}

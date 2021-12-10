package admin

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

func link(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	remoteID, err := node.Contacts.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	_, err = node.Linker.Connect(ctx, node.Peer(remoteID))
	if err != nil {
		return err
	}

	return nil
}

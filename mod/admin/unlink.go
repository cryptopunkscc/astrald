package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func unlink(w io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	identity, err := node.Contacts.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	peer := node.Peers.Find(identity)
	if peer == nil {
		return errors.New("peer not found")
	}

	peer.Unlink()

	return nil
}

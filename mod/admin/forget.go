package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func forget(w io.ReadWriter, node *node.Node, args []string) error {
	if len(args) == 0 {
		return errors.New("missing node id")
	}

	identity, err := node.Contacts.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	if err := node.Tracker.ForgetIdentity(identity); err != nil {
		return err
	}

	return node.Contacts.Delete(identity)
}

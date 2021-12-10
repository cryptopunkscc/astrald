package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"io"
)

func add(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := contacts.ParseInfo(args[0])
	if err != nil {
		return err
	}

	node.Contacts.AddInfo(r)

	return nil
}

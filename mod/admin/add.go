package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func add(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	return node.Contacts.AddNodeInfo(args[0])
}

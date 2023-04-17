package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
)

func add(_ io.ReadWriter, node node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	nodeInfo, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	if nodeInfo.Identity.IsEqual(node.Identity()) {
		return errors.New("cannot add self")
	}

	nodeinfo.AddToTracker(nodeInfo, node.Tracker())
	return nodeinfo.AddToContacts(nodeInfo, node.Contacts())
}

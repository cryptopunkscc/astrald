package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func cmdTracker(w io.ReadWriter, node node.Node, _ []string) error {
	ids, err := node.Tracker().Identities()
	if err != nil {
		return err
	}

	for _, nodeID := range ids {
		fmt.Fprintln(w, nodeID.PublicKeyHex())

		addrs, err := node.Tracker().AddrByIdentity(nodeID)
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			printEndpoint(w, addr)
		}

		fmt.Fprintln(w)
	}

	return nil
}

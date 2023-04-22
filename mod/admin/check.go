package admin

import (
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

// check forces a health check of all links of all peers
func check(w io.ReadWriter, node node.Node, args []string) error {
	for _, peer := range node.Network().Peers().All() {
		peer.Check()
	}
	return nil
}

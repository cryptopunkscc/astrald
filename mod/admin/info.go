package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func info(w io.ReadWriter, node *node.Node, _ []string) error {
	fmt.Fprintln(w, "nodeID   ", node.Identity())
	fmt.Fprintln(w, "alias    ", node.Alias())
	fmt.Fprintln(w, "nodeinfo ", node.NodeInfo())
	for _, addr := range node.Infra.Addresses() {
		printAddr(w, addr)
	}
	return nil
}

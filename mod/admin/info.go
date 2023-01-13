package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
)

func info(w io.ReadWriter, node *node.Node, _ []string) error {
	nodeInfo := nodeinfo.New(node.Identity())
	nodeInfo.Alias = node.Alias()

	for _, a := range node.Infra.Addresses() {
		if a.Global {
			nodeInfo.Addresses = append(nodeInfo.Addresses, a.Addr)
		}
	}

	fmt.Fprintln(w, "nodeID   ", node.Identity())
	fmt.Fprintln(w, "alias    ", node.Alias())
	fmt.Fprintln(w, "nodeinfo ", nodeInfo)
	for _, addr := range node.Infra.Addresses() {
		printAddr(w, addr)
	}
	return nil
}

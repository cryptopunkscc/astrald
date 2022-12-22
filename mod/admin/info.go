package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
)

func info(w io.ReadWriter, node *node.Node, _ []string) error {
	i := nodeinfo.New(node.Identity())
	i.Alias = node.Alias()

	for _, a := range node.Infra.Addresses() {
		if a.Global {
			i.Addresses = append(i.Addresses, a.Addr)
		}
	}

	fmt.Fprintln(w, "nodeID   ", node.Identity())
	fmt.Fprintln(w, "alias    ", node.Alias())
	fmt.Fprintln(w, "nodeinfo ", i)
	for _, addr := range node.Infra.Addresses() {
		printAddr(w, addr)
	}
	return nil
}

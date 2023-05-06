package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
)

func info(w io.ReadWriter, node node.Node, _ []string) error {
	nodeInfo := nodeinfo.FromNode(node)

	fmt.Fprintln(w, "nodeID   ", node.Identity())
	fmt.Fprintln(w, "alias    ", node.Alias())
	fmt.Fprintln(w, "nodeinfo ", nodeInfo)
	for _, e := range node.Infra().Endpoints() {
		e, _ := node.Infra().Unpack(e.Network(), e.Pack())
		printEndpoint(w, e)
	}
	return nil
}

package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func cmdContacts(w io.ReadWriter, node *node.Node, _ []string) error {
	for nodeID := range node.Contacts.Identities() {
		fmt.Fprintln(w, "node", nodeID.PublicKeyHex())
		fmt.Fprintln(w, "alias", node.Contacts.GetAlias(nodeID))

		for addr := range node.Contacts.Resolve(nodeID) {
			printAddr(w, addr)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func printAddr(w io.Writer, addr infra.Addr) (int, error) {
	return fmt.Fprintf(w, "  %-10s%s\n", addr.Network(), addr.String())
}

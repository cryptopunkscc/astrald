package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func present(w io.ReadWriter, node *node.Node, _ []string) error {
	for i := range node.Presence.Identities() {
		fmt.Fprintln(w, node.Contacts.DisplayName(i))
	}
	return nil
}

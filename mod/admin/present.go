package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func present(w io.ReadWriter, node *node.Node, args []string) error {
	for i := range node.Presence.Identities() {
		fmt.Fprintln(w, i.String())
	}
	return nil
}

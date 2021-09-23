package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/logfmt"
	_node "github.com/cryptopunkscc/astrald/node"
	"io"
)

func links(stream io.ReadWriter, node *_node.Node, _ []string) error {
	for _, link := range node.LinkCache.Links() {
		fmt.Fprintln(stream,
			link.RemoteIdentity().String(),
			logfmt.Dir(link.Outbound()),
			link.RemoteAddr().Network(),
			link.RemoteAddr().String(),
		)
	}
	return nil
}

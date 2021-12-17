package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

func peers(w io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Peers() {
		peerName := node.Contacts.DisplayName(peer.Identity())

		fmt.Fprintf(w, "peer %s (idle %s)\n",
			peerName,
			peer.Idle().Round(time.Second),
		)
		for link := range peer.Links() {
			remoteAddr := link.RemoteAddr().String()

			if gwAddr, ok := link.RemoteAddr().(gw.Addr); ok {
				remoteAddr = "via " + node.Contacts.DisplayName(gwAddr.Gate())
			}

			fmt.Fprintf(w,
				"  %s %s %s (idle %s, lat %.1fms)\n",
				logfmt.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				remoteAddr,
				link.Idle().Round(time.Second),
				float64(link.Latency().Microseconds())/1000,
			)
			for c := range link.Conns() {
				fmt.Fprintf(w,
					"    %s %s (idle %s)\n",
					logfmt.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					c.Idle().Round(time.Second),
				)
			}
		}
	}

	return nil
}

package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

func peers(w io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Peers.All(nil) {
		peerName := node.Contacts.DisplayName(peer.Identity())

		fmt.Fprintf(w, "peer %s (idle %s)\n",
			peerName,
			peer.Idle().Round(time.Second),
		)
		for _, link := range peer.Links() {
			remoteAddr := link.RemoteAddr().String()

			if gwAddr, ok := link.RemoteAddr().(gw.Addr); ok {
				remoteAddr = "via " + node.Contacts.DisplayName(gwAddr.Gate())
			}

			fmt.Fprintf(w,
				"  %s %s %s (idle %s, age %s, prio %d, ping %.1fms)\n",
				log.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				remoteAddr,
				link.Idle().Round(time.Second),
				time.Since(link.EstablishedAt()).Round(time.Second),
				link.Priority(),
				float64(link.Ping().Microseconds())/1000,
			)
			for c := range link.Conns() {
				fmt.Fprintf(w,
					"    %s %s [%d:%d] (idle %s)\n",
					log.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					c.InputStream().ID(),
					c.OutputStream().ID(),
					c.Idle().Round(time.Second),
				)
			}
		}

		fmt.Fprintf(w, "\n")
	}

	for l := range node.Peers.Linkers() {
		peerName := node.Contacts.DisplayName(l)
		fmt.Fprintf(w, "peer %s linking...\n", peerName)
	}

	return nil
}

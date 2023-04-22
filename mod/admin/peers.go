package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network"
	"io"
	"time"
)

func peers(w io.ReadWriter, node node.Node, _ []string) error {
	for _, peer := range node.Network().Peers().All() {
		peerName := node.Contacts().DisplayName(peer.Identity())

		fmt.Fprintf(w, "peer %s (idle %s)\n",
			peerName,
			peer.Idle().Round(time.Second),
		)
		for _, link := range peer.Links() {
			remoteAddr := link.RemoteAddr().String()

			if gwAddr, ok := link.RemoteAddr().(gw.Addr); ok {
				remoteAddr = "via " + node.Contacts().DisplayName(gwAddr.Gate())
			}

			fmt.Fprintf(w,
				"  %s %s %s (idle %s, age %s, prio %d, ping %.1fms)\n",
				log.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				remoteAddr,
				link.Activity().Idle().Round(time.Second),
				time.Since(link.EstablishedAt()).Round(time.Second),
				link.Priority(),
				float64(link.Health().AverageRTT().Microseconds())/1000,
			)
			for _, c := range link.Conns().All() {
				fmt.Fprintf(w,
					"    %s %s [%d:%d] (idle %s)\n",
					log.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					c.LocalPort(),
					c.RemotePort(),
					c.Idle().Round(time.Second),
				)
			}
		}

		fmt.Fprintf(w, "\n")
	}

	if n, ok := node.Network().(*network.Network); ok {
		for _, l := range n.Linkers() {
			peerName := node.Contacts().DisplayName(l)
			fmt.Fprintf(w, "peer %s linking...\n", peerName)
		}
	}

	return nil
}

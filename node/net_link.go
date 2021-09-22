package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"time"
)

func PeerLink(nodeID *id.Identity, peerInfo *PeerInfo, network *Network) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		var wg sync.WaitGroup

		for _, netName := range net.UnicastNetworks() {
			wg.Add(1)
			go func(netName string) {
				defer wg.Done()

				for lnk := range NetLink(nodeID, netName, peerInfo, network) {
					outCh <- lnk
				}
			}(netName)
		}

		wg.Wait()
	}()

	return outCh
}

func NetLink(nodeID *id.Identity, netName string, peerInfo *PeerInfo, network *Network) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		for {
			addrs := peerInfo.NodeAddr(nodeID.String()).OnlyNet(netName)

			for _, addr := range addrs {
				lnk, err := network.LinkAt(nodeID, addr)
				if err != nil {
					continue
				}
				outCh <- lnk
				<-lnk.WaitClose()
				break
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return outCh
}

package link

import "github.com/cryptopunkscc/astrald/auth/id"

type Filter <-chan *Link

func (filter Filter) Select(selector func(*Link) bool) Filter {
	selected := make(chan *Link)

	go func() {
		defer close(selected)

		for link := range filter {
			if selector(link) {
				selected <- link
			}
		}
	}()

	return selected
}

func (filter Filter) Peer(peerID id.Identity) Filter {
	return filter.Select(func(l *Link) bool {
		return l.RemoteIdentity().IsEqual(peerID)
	})
}

func (filter Filter) Network(networkName string) Filter {
	return filter.Select(func(l *Link) bool {
		return l.RemoteAddr().Network() == networkName //TODO: remote addr is sometimes unknown, but the network always is
	})
}

func (filter Filter) Outbound(outbound bool) Filter {
	return filter.Select(func(l *Link) bool {
		return l.Outbound() == outbound
	})
}

func (filter Filter) Count() (n int) {
	for _ = range filter {
		n++
	}
	return
}

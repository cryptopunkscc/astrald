package link

import "github.com/cryptopunkscc/astrald/auth/id"

type Stream <-chan *Link

func (stream Stream) Select(selector func(*Link) bool) Stream {
	selected := make(chan *Link)

	go func() {
		defer close(selected)

		for link := range stream {
			if selector(link) {
				selected <- link
			}
		}
	}()

	return selected
}

func (stream Stream) Peer(peerID id.Identity) Stream {
	return stream.Select(func(l *Link) bool {
		return l.RemoteIdentity().IsEqual(peerID)
	})
}

func (stream Stream) Network(networkName string) Stream {
	return stream.Select(func(l *Link) bool {
		return l.RemoteAddr().Network() == networkName //TODO: remote addr is sometimes unknown, but the network always is
	})
}

func (stream Stream) Outbound(outbound bool) Stream {
	return stream.Select(func(l *Link) bool {
		return l.Outbound() == outbound
	})
}

func (stream Stream) Count() (n int) {
	for _ = range stream {
		n++
	}
	return
}

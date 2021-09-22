package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	_link "github.com/cryptopunkscc/astrald/node/link"
)

type LinkPool []*_link.Link

func (pool LinkPool) Select(slectFunc func(*_link.Link) bool) LinkPool {
	list := make(LinkPool, 0)

	for _, link := range pool {
		if slectFunc(link) {
			list = append(list, link)
		}
	}

	return list
}

func (pool LinkPool) OnlyPeer(peerID *id.Identity) LinkPool {
	return pool.Select(func(l *_link.Link) bool {
		return l.RemoteIdentity().String() == peerID.String()
	})
}

func (pool LinkPool) OnlyNet(netName string) LinkPool {
	return pool.Select(func(l *_link.Link) bool {
		return l.RemoteAddr().Network() == netName
	})
}

func (pool LinkPool) Outbound(out bool) LinkPool {
	return pool.Select(func(l *_link.Link) bool {
		return l.Outbound() == out
	})
}

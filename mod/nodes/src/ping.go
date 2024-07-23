package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"time"
)

const pingTimeout = time.Second * 30

func (mod *Module) pingStream(s *Stream) time.Duration {
	var nonce = astral.NewNonce()
	var p = &Ping{
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	}

	_, ok := mod.pings.Set(nonce, p)
	if !ok {
		return 0
	}
	defer mod.pings.Delete(nonce)

	err := s.Write(&frames.Ping{
		Nonce: nonce,
	})
	if err != nil {
		return -1
	}
	p.sentAt = time.Now()

	select {
	case <-p.pong:
		return time.Since(p.sentAt)
	case <-time.After(pingTimeout):
		return -1
	}
}

func (mod *Module) handlePing(source *Stream, ping *frames.Ping) {
	if ping.Pong {
		p, ok := mod.pings.Delete(ping.Nonce)
		if !ok {
			mod.log.Errorv(1, "invalid pong nonce from %v", source.RemoteIdentity())
			return
		}
		close(p.pong)
	} else {
		source.Write(&frames.Ping{
			Nonce: ping.Nonce,
			Pong:  true,
		})
	}
}

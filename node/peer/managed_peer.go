package peer

import (
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type managedPeer struct {
	wg   sync.WaitGroup
	peer *Peer
}

func (mp *managedPeer) hold(done sig.Signal) {
	mp.wg.Add(1)

	sig.On(done, mp.wg.Done)
}

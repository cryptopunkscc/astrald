package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type Linker struct {
	remoteID id.Identity
	err      error
	link     net.Link
	done     chan struct{}
}

func (l *Linker) RemoteID() id.Identity {
	return l.remoteID
}

func (l *Linker) Done() <-chan struct{} {
	return l.done
}

func (l *Linker) Error() error {
	return l.err
}

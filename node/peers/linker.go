package peers

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Linker struct {
	remoteID id.Identity
	err      error
	link     *link.Link
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

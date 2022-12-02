package peers

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Linker struct {
	remoteID id.Identity
	ctx      context.Context
	err      error
}

func (l *Linker) RemoteID() id.Identity {
	return l.remoteID
}

func (l *Linker) Done() <-chan struct{} {
	return l.ctx.Done()
}

func (l *Linker) Error() error {
	return l.err
}

package nodes

import (
	"context"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

const upgradeTimeout = 3 * time.Minute

func (mod *Module) connectivityUpgrade(e *nodes.StreamPressureEvent) {
	sw := &sig.Switch{}
	if existing, ok := mod.upgraders.Set(e.RemoteIdentity.String(), sw); !ok {
		sw = existing
	}

	sw.Run(mod.ctx, func(_ context.Context) {
		ctx, cancel := mod.ctx.WithTimeout(upgradeTimeout)
		defer cancel()

		mod.log.Log("connectivity upgrade: starting with %v", e.RemoteIdentity)

		result := <-mod.linkPool.RetrieveLink(ctx, e.RemoteIdentity,
			WithForceNew(), WithStrategies(nodes.StrategyNAT))

		if result.Err != nil {
			mod.log.Log("connectivity upgrade with %v failed: %v", e.RemoteIdentity, result.Err)
			return
		}

		mod.log.Log("connectivity upgrade: new stream %v (%v)", result.Stream.id, result.Stream.Network())

		mod.migrateSessions(e.StreamID, result.Stream)
	})
}

func (mod *Module) migrateSessions(oldStreamID astral.Nonce, newStream *Stream) {
	// todo: migrate sessions from oldStreamID to newStream
}

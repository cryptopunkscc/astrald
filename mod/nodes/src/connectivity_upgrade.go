package nodes

import (
	"context"
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodesClient "github.com/cryptopunkscc/astrald/mod/nodes/client"
	"github.com/cryptopunkscc/astrald/sig"
)

const upgradeTimeout = 3 * time.Minute
const upgradeCooldown = 5 * time.Minute

func (mod *Module) connectivityUpgrade(e *nodes.StreamPressureEvent) {
	connectivityGate := &sig.Switch{}
	if existing, ok := mod.upgraders.Set(e.RemoteIdentity.String(), connectivityGate); !ok {
		connectivityGate = existing
	}

	connectivityGate.Run(mod.ctx, func(_ context.Context) {
		var targetStream *Stream

		// prefer an existing stream over establishing a new one
		alternatives := mod.peers.streams.Select(func(s *Stream) bool {
			return s.RemoteIdentity().IsEqual(e.RemoteIdentity) && s.id != e.StreamID
		})
		slices.SortFunc(alternatives, func(a, b *Stream) int {
			if (a.pressure == nil) == (b.pressure == nil) {
				return 0
			}
			if a.pressure == nil {
				return -1
			}
			return 1
		})

		if len(alternatives) > 0 {
			targetStream = alternatives[0]
			mod.log.Log("connectivity upgrade: reusing existing stream %v (%v)", targetStream.id, targetStream.Network())
		} else {
			ctx, cancel := mod.ctx.WithTimeout(upgradeTimeout)
			defer cancel()

			mod.log.Log("connectivity upgrade: starting NAT traversal with %v", e.RemoteIdentity)

			result := <-mod.linkPool.RetrieveLink(ctx, e.RemoteIdentity,
				WithForceNew(), WithStrategies(nodes.StrategyNAT))

			if result.Err != nil {
				mod.log.Log("connectivity upgrade with %v failed: %v", e.RemoteIdentity, result.Err)
			} else {
				targetStream = result.Stream
				mod.log.Log("connectivity upgrade: new stream %v (%v)", targetStream.id, targetStream.Network())
			}
		}

		if targetStream != nil {
			mod.migrateSessions(e.StreamID, targetStream)
		}

		select {
		case <-time.After(upgradeCooldown):
		case <-mod.ctx.Done():
		}
	})
}

func (mod *Module) migrateSessions(oldStreamID astral.Nonce, newStream *Stream) {
	sessions := mod.peers.sessions.Select(func(_ astral.Nonce, s *session) bool {
		return s.stream != nil && s.stream.id == oldStreamID && s.IsOpen()
	})

	if len(sessions) == 0 {
		return
	}

	client := nodesClient.New(mod.node.Identity(), astrald.Default())
	for _, session := range sessions {
		if err := client.StartSessionMigration(mod.ctx, session.Nonce, newStream.id); err != nil {
			mod.log.Error("migrate session %v failed: %v", session.Nonce, err)
		}
	}
}

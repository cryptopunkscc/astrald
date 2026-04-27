package nodes

import (
	"context"
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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
		mod.log.Log("connectivity upgrade triggered for %v (stream %v)", e.RemoteIdentity, e.StreamID)

		var targetStream *Stream
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
				mod.log.Log("connectivity upgrade: established new stream %v (%v)", targetStream.id, targetStream.Network())
			}
		}

		if targetStream != nil {
			mod.migrateSessions(e.StreamID, targetStream)
		} else {
			mod.log.Logv(2, "connectivity upgrade: no target stream found, skipping migration for %v", e.RemoteIdentity)
		}

		select {
		case <-time.After(upgradeCooldown):
		case <-mod.ctx.Done():
		}
	})
}

const migrateSessionTimeout = 30 * time.Second

func (mod *Module) migrateSessions(oldStreamID astral.Nonce, newStream *Stream) {
	oldStream := mod.findStreamByID(oldStreamID)
	if oldStream == nil {
		mod.log.Logv(1, "migrate sessions: old stream %v not found", oldStreamID)
		return
	}

	sessions := mod.peers.sessions.Select(func(k astral.Nonce, v *session) bool {
		return v.IsOpen() && v.isOnStream(oldStream) && v.CanAutoMigrate()
	})

	if len(sessions) == 0 {
		return
	}

	mod.log.Log("migrating %v sessions from stream %v to %v", len(sessions), oldStreamID, newStream.id)

	var migrated int
	for _, sess := range sessions {
		ctx, cancel := mod.ctx.WithTimeout(migrateSessionTimeout)
		err := mod.migrateSession(ctx, sess, newStream)
		cancel()

		if err != nil {
			mod.log.Logv(1, "migrate session %v failed: %v", sess.Nonce, err)
			continue
		}
		migrated++
	}

	mod.log.Log("migrated %v/%v sessions from stream %v to %v", migrated, len(sessions), oldStreamID, newStream.id)
}

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

func (mod *Module) connectivityUpgrade(e *nodes.LinkPressureEvent) {
	connectivityGate := &sig.Switch{}
	if existing, ok := mod.upgraders.Set(e.RemoteIdentity.String(), connectivityGate); !ok {
		connectivityGate = existing
	}

	connectivityGate.Run(mod.ctx, func(_ context.Context) {
		mod.log.Log("connectivity upgrade triggered for %v (link %v)", e.RemoteIdentity, e.LinkID)

		var targetLink *Link
		var alternatives []*Link
		for _, s := range mod.linkPool.Links().Values() {
			if s.RemoteIdentity().IsEqual(e.RemoteIdentity) && s.ID() != e.LinkID {
				alternatives = append(alternatives, s)
			}
		}

		slices.SortFunc(alternatives, func(a, b *Link) int {
			if a.IsHighPressure() && !b.IsHighPressure() {
				return -1
			}

			if !a.IsHighPressure() && b.IsHighPressure() {
				return 1
			}

			return 0
		})

		if len(alternatives) > 0 {
			targetLink = alternatives[0]
			mod.log.Log("connectivity upgrade: reusing existing link %v (%v)", targetLink.ID(), targetLink.Network())
		} else {
			ctx, cancel := mod.ctx.WithTimeout(upgradeTimeout)
			defer cancel()

			mod.log.Log("connectivity upgrade: starting NAT traversal with %v", e.RemoteIdentity)
			result := <-mod.linkPool.RetrieveLink(ctx, e.RemoteIdentity,
				WithForceNew(), WithStrategies(nodes.StrategyNAT))

			if result.Err != nil {
				mod.log.Log("connectivity upgrade with %v failed: %v", e.RemoteIdentity, result.Err)
			} else {
				targetLink = result.Link
				mod.log.Log("connectivity upgrade: established new link %v (%v)", targetLink.ID(), targetLink.Network())
			}
		}

		if targetLink != nil {
			mod.migrateSessions(e.LinkID, targetLink)
		} else {
			mod.log.Logv(2, "connectivity upgrade: no target link found, skipping migration for %v", e.RemoteIdentity)
		}

		select {
		case <-time.After(upgradeCooldown):
		case <-mod.ctx.Done():
		}
	})
}

const migrateSessionTimeout = 30 * time.Second

func (mod *Module) migrateSessions(oldLinkID astral.Nonce, newLink *Link) {
	oldLink := mod.findLinkByID(oldLinkID)
	if oldLink == nil {
		mod.log.Logv(1, "migrate sessions: old link %v not found", oldLinkID)
		return
	}

	sessions := oldLink.Mux.sessions.Select(func(k astral.Nonce, v *session) bool {
		return v.IsOpen() && v.CanAutoMigrate()
	})

	if len(sessions) == 0 {
		return
	}

	mod.log.Log("migrating %v sessions from link %v to %v", len(sessions), oldLinkID, newLink.ID())

	var migrated int
	for _, sess := range sessions {
		ctx, cancel := mod.ctx.WithTimeout(migrateSessionTimeout)
		err := mod.migrateSession(ctx, sess, newLink)
		cancel()

		if err != nil {
			mod.log.Logv(1, "migrate session %v failed: %v", sess.Nonce, err)
			continue
		}
		migrated++
	}

	mod.log.Log("migrated %v/%v sessions from link %v to %v", migrated, len(sessions), oldLinkID, newLink.ID())
}

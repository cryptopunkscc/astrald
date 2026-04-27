package nodes

import (
	"io"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodescli "github.com/cryptopunkscc/astrald/mod/nodes/client"
)

const relayIdleTimeout = 30 * time.Minute

var _ astral.Router = &relayChannel{}

type relayChannel struct {
	mod      *Module          // for session registry access and self-removal
	relayID  *astral.Identity // relay node identity (transport, not the actual target)
	ch       *channel.Channel
	mu       sync.Mutex
	lastUsed time.Time
}

func (rc *relayChannel) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	if !ctx.Identity().IsEqual(q.Caller) {
		if err := rc.mod.sendCallerProof(ctx, q, rc.relayID); err != nil {
			return query.RouteNotFound()
		}
	}

	conn, ok := rc.mod.peers.sessions.Set(q.Nonce, newSession(q.Nonce))
	if !ok {
		return query.RouteNotFound()
	}

	conn.RemoteIdentity = q.Target
	conn.relayID = rc.relayID
	conn.Query = q.QueryString
	conn.Outbound = true

	container := nodes.NewQueryContainer(q, defaultBufferSize)

	rc.mu.Lock()
	err := rc.ch.Send(container)
	if err == nil {
		rc.lastUsed = time.Now()
	}
	rc.mu.Unlock()

	if err != nil {
		rc.mod.relayChannels.Delete(rc.relayID.String()) // channel is broken, evict
		conn.Close()
		return query.RouteNotFound()
	}

	select {
	case errCode := <-conn.res:
		if errCode != 0 {
			return query.RejectWithCode(errCode)
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()
		return conn, nil

	case <-ctx.Done():
		conn.Close()
		return query.RouteNotFound()
	}
}

func (mod *Module) newRelayChannel(relayID *astral.Identity, ch *channel.Channel) *relayChannel {
	return &relayChannel{mod: mod, relayID: relayID, ch: ch, lastUsed: time.Now()}
}

func (mod *Module) getRelay(ctx *astral.Context, relayID *astral.Identity) (*relayChannel, error) {
	key := relayID.String()
	rc, ok := mod.relayChannels.Get(key)
	if ok {
		return rc, nil
	}

	ch, err := nodescli.Default().WithTarget(relayID).OpenRelay(ctx)
	if err != nil {
		return nil, err
	}

	rc, ok = mod.relayChannels.Set(key, mod.newRelayChannel(relayID, ch))
	if !ok {
		ch.Close() // race lost — use existing
	} else {
		go rc.watch()
	}

	return rc, nil
}

func (rc *relayChannel) watch() {
	mod := rc.mod
	timer := time.NewTimer(relayIdleTimeout)
	defer timer.Stop()

	for {
		select {
		case <-mod.ctx.Done():
			rc.mu.Lock()
			rc.ch.Close()
			rc.mu.Unlock()
			return
		case <-timer.C:
			rc.mu.Lock()
			idle := time.Since(rc.lastUsed)
			rc.mu.Unlock()

			if idle < relayIdleTimeout {
				timer.Reset(relayIdleTimeout - idle)
				continue
			}

			rc.mu.Lock()
			mod.relayChannels.Delete(rc.relayID.String())
			rc.ch.Close()
			rc.mu.Unlock()
			return
		}
	}
}

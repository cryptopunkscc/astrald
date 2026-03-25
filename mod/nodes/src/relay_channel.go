package nodes

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodescli "github.com/cryptopunkscc/astrald/mod/nodes/client"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

const relayIdleTimeout = 30 * time.Minute

type relayChannel struct {
	ch       *channel.Channel
	mu       sync.Mutex
	lastUsed atomic.Int64 // unix nano
}

func (rc *relayChannel) send(container *nodes.QueryContainer) (err error) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if err := rc.ch.Send(container); err != nil {
		return err
	}

	defer func() {
		if err == nil {
			rc.lastUsed.Store(time.Now().UnixNano())
		}
	}()

	return rc.ch.Switch(channel.ExpectAck, channel.PassErrors)
}

func (mod *Module) relayQuery(ctx *astral.Context, q *astral.Query, relayID *astral.Identity, w io.WriteCloser) (io.WriteCloser, error) {
	if !ctx.Identity().IsEqual(q.Caller) {
		if err := mod.sendCallerProof(ctx, q, relayID); err != nil {
			return query.RouteNotFound(mod, fmt.Errorf("caller proof: %w", err))
		}
	}

	conn, ok := mod.peers.sessions.Set(q.Nonce, newSession(q.Nonce))
	if !ok {
		return query.RouteNotFound(mod, errors.New("session nonce already in use"))
	}
	conn.RemoteIdentity = relayID
	conn.Query = q.Query
	conn.Outbound = true

	key := relayID.String()
	rc, ok := mod.relayChannels.Get(key)
	if !ok {
		ch, err := nodescli.Default().WithTarget(relayID).OpenRelay(ctx)
		if err != nil {
			conn.swapState(stateRouting, stateClosed)
			mod.peers.sessions.Delete(q.Nonce)
			return query.RouteNotFound(mod, err)
		}

		newRC := &relayChannel{ch: ch}
		newRC.lastUsed.Store(time.Now().UnixNano())

		rc, ok = mod.relayChannels.Set(key, newRC)
		if !ok {
			ch.Close() // race lost — use existing
		} else {
			go mod.watchRelayChannel(key, newRC)
		}
	}

	container := &nodes.QueryContainer{
		CallerID: q.Caller,
		TargetID: q.Target,
		Query: frames.Query{
			Nonce:  q.Nonce,
			Buffer: uint32(conn.rsize),
			Query:  q.Query,
		},
	}

	if err := rc.send(container); err != nil {
		mod.relayChannels.Delete(key)
		conn.swapState(stateRouting, stateClosed)
		mod.peers.sessions.Delete(q.Nonce)
		return query.RouteNotFound(mod, fmt.Errorf("send relay container: %w", err))
	}

	select {
	case errCode := <-conn.res:
		if errCode != 0 {
			mod.peers.sessions.Delete(q.Nonce)
			return query.RejectWithCode(errCode)
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.swapState(stateRouting, stateClosed)
		mod.peers.sessions.Delete(q.Nonce)
		return query.RouteNotFound(mod, ctx.Err())
	}
}

func (mod *Module) watchRelayChannel(key string, rc *relayChannel) {
	timer := time.NewTimer(relayIdleTimeout)
	defer timer.Stop()

	for {
		select {
		case <-mod.ctx.Done():
			rc.ch.Close()
			return
		case <-timer.C:
			idle := time.Since(time.Unix(0, rc.lastUsed.Load()))
			if idle < relayIdleTimeout {
				timer.Reset(relayIdleTimeout - idle)
				continue
			}

			rc.mu.Lock()
			mod.relayChannels.Delete(key)
			rc.ch.Close()
			rc.mu.Unlock()
			return
		}
	}
}

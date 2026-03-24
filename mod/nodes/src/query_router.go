package nodes

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodescli "github.com/cryptopunkscc/astrald/mod/nodes/client"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
)

func (mod *Module) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (rw io.WriteCloser, err error) {
	// check if the context allows for network queries
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return query.RouteNotFound(mod, astral.ErrZoneExcluded)
	}

	// check if we're querying ourselves
	if q.Target.IsEqual(mod.node.Identity()) {
		return query.RouteNotFound(mod)
	}

	// are we linked already?
	if mod.IsPeer(q.Target) {
		if !ctx.Identity().IsEqual(q.Caller) {
			if err := mod.sendCallerProof(ctx, q, q.Target); err != nil {
				return query.RouteNotFound(mod, err)
			}
		}
		return mod.peers.RouteQuery(ctx, q, w)
	}

	retrieveCtx, cancel := ctx.WithTimeout(120 * time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return query.RouteNotFound(mod, ctx.Err())
	case result := <-mod.linkPool.RetrieveLink(retrieveCtx, q.Target, WithStrategies(nodes.StrategyBasic, nodes.StrategyTor)):
		if result.Err != nil {
			mod.log.Error("retrieve link failed: %v", result.Err)
			break
		}

		if !ctx.Identity().IsEqual(q.Caller) {
			if err := mod.sendCallerProof(ctx, q, q.Target); err != nil {
				return query.RouteNotFound(mod, err)
			}
		}
		return mod.peers.RouteQuery(ctx, q, w)
	}

	// try relays
	relayValue, ok := q.Extra.Get(nodes.ExtraRelayVia)
	if !ok {
		return query.RouteNotFound(mod)
	}

	relays, ok := relayValue.([]*astral.Identity)
	if !ok {
		return query.RouteNotFound(mod)
	}

	for _, relayID := range relays {
		// never use the target as a relay to itself
		if relayID.IsEqual(q.Target) {
			continue
		}

		rw, err := mod.routeViaRelay(ctx, q, relayID, w)
		if err == nil {
			return rw, nil
		}
	}

	return query.RouteNotFound(mod)
}

func (mod *Module) routeViaRelay(ctx *astral.Context, q *astral.Query, relayID *astral.Identity, w io.WriteCloser) (io.WriteCloser, error) {
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

	// build and send the relay container
	container := &nodes.QueryContainer{
		CallerID: q.Caller,
		TargetID: q.Target,
		Query: frames.Query{
			Nonce:  q.Nonce,
			Buffer: uint32(conn.rsize),
			Query:  q.Query,
		},
	}

	if err := nodescli.New(relayID, nil).SendRelayedQuery(ctx, container); err != nil {
		conn.swapState(stateRouting, stateClosed)
		mod.peers.sessions.Delete(q.Nonce)
		return query.RouteNotFound(mod, fmt.Errorf("send relay container: %w", err))
	}

	// wait for frames.Response from the relay peer stream
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

func (mod *Module) sendCallerProof(ctx *astral.Context, q *astral.Query, target *astral.Identity) error {
	v, ok := q.Extra.Get(nodes.ExtraCallerProof)
	if !ok {
		return errors.New("missing caller proof")
	}

	callerProof := v.(astral.Object)
	if callerProof == nil {
		return errors.New("missing caller proof")
	}

	err := mod.Objects.Push(ctx, target, callerProof)
	if err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

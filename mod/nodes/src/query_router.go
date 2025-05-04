package nodes

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"io"
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
		return mod.peers.RouteQuery(ctx, q, w)
	}

	// try to link
	ch, err := mod.ResolveEndpoints(ctx, q.Target)
	if err == nil {
		err = mod.peers.connectAtAny(ctx, q.Target, ch)
		if err == nil {
			return mod.peers.RouteQuery(ctx, q, w)
		}
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

		// try to configure the relay
		err := mod.configureRelay(ctx, q, relayID)
		if err != nil {
			continue
		}

		// relay the query
		var rq = &astral.Query{
			Nonce:  q.Nonce,
			Caller: q.Caller,
			Target: relayID,
			Query:  q.Query,
			Extra:  *q.Extra.Copy(),
		}

		w, err := mod.peers.RouteQuery(ctx, rq, w)
		if err == nil {
			return w, nil
		}
	}

	return query.RouteNotFound(mod)
}

func (mod *Module) configureRelay(ctx *astral.Context, q *astral.Query, relayID *astral.Identity) error {
	var caller, target *astral.Identity

	// check if we need to change the caller
	if !ctx.Identity().IsEqual(q.Caller) {
		err := mod.sendCallerProof(ctx, q, relayID)
		if err != nil {
			return fmt.Errorf("send caller proof: %w", err)
		}
		caller = q.Caller
	}

	// check if we need to change the target
	if !relayID.IsEqual(q.Target) {
		target = q.Target
	}

	// return if no changes are required
	if caller == nil && target == nil {
		return nil
	}

	// configure the relay
	err := mod.on(relayID).Relay(ctx, q.Nonce, caller, target)
	if err != nil {
		return fmt.Errorf("relay query: %w", err)
	}

	return nil
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

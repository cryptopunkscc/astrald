package nodes

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (rw io.WriteCloser, err error) {
	// check if the context allows for network queries
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return query.RouteNotFound()
	}

	// check if we're querying ourselves
	if q.Target.IsEqual(mod.node.Identity()) {
		return query.RouteNotFound()
	}

	if mod.IsPeer(q.Target) {
		// if we are not the caller it means we are a relay for identity.
		if !ctx.Identity().IsEqual(q.Caller) {
			rc, err := mod.getRelay(ctx, q.Target)
			if err != nil {
				return query.RouteNotFound()
			}
			return rc.RouteQuery(ctx, q, w)
		}

		return mod.peers.RouteQuery(ctx, q, w)
	}

	retrieveCtx, cancel := ctx.WithTimeout(120 * time.Second)
	defer cancel()

	// todo: there is error printed out  when calling identity that we cannot link with (e.g. other's node app)
	select {
	case <-ctx.Done():
		return query.RouteNotFound()
	case result := <-mod.linkPool.RetrieveLink(retrieveCtx, q.Target, WithStrategies(nodes.StrategyBasic, nodes.StrategyTor)):
		if result.Err != nil {
			mod.log.Error("retrieve link failed: %v", result.Err)
			break
		}

		if !ctx.Identity().IsEqual(q.Caller) {
			rc, err := mod.getRelay(ctx, q.Target)
			if err != nil {
				return query.RouteNotFound()
			}
			return rc.RouteQuery(ctx, q, w)
		}

		return mod.peers.RouteQuery(ctx, q, w)
	}

	// try relays
	relayValue, ok := q.Extra.Get(nodes.ExtraRelayVia)
	if !ok {
		return query.RouteNotFound()
	}

	relays, ok := relayValue.([]*astral.Identity)
	if !ok {
		return query.RouteNotFound()
	}

	for _, relayID := range relays {
		// never use the target as a relay to itself
		if relayID.IsEqual(q.Target) {
			continue
		}

		rc, err := mod.getRelay(ctx, relayID)
		if err != nil {
			continue
		}

		rw, err := rc.RouteQuery(ctx, q, w)
		if err == nil {
			return rw, nil
		}
	}

	return query.RouteNotFound()
}

func (mod *Module) sendCallerProof(ctx *astral.Context, q *astral.InFlightQuery, target *astral.Identity) error {
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

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
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return query.RouteNotFound()
	}

	if q.Target.IsEqual(mod.node.Identity()) {
		return query.RouteNotFound()
	}

	if link := mod.linkPool.SelectLinkWith(q.Target); link != nil {
		return link.RouteQuery(ctx, q, w)
	}

	retrieveCtx, cancel := ctx.WithTimeout(120 * time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return query.RouteNotFound()
	case result := <-mod.linkPool.RetrieveLink(retrieveCtx, q.Target, WithStrategies(nodes.StrategyBasic, nodes.StrategyTor)):
		if result.Err != nil {
			mod.log.Error("retrieve link failed: %v", result.Err)
			break
		}
		return result.Stream.RouteQuery(ctx, q, w)
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
		if relayID.IsEqual(q.Target) {
			continue
		}

		relayLink := mod.linkPool.SelectLinkWith(relayID)
		if relayLink == nil {
			continue
		}

		if !ctx.Identity().IsEqual(q.Caller) {
			if err := mod.sendCallerProof(ctx, q, relayID); err != nil {
				continue
			}
		}

		rw, err := relayLink.RouteQuery(ctx, q, w)
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

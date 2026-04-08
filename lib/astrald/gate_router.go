package astrald

import "github.com/cryptopunkscc/astrald/astral"

// ReadyGate Ready() returns a channel that is closed when the gate is open and replaced
// with a new open channel when the gate closes (session dropped).
type ReadyGate interface {
	Ready() <-chan struct{}
}

type GateRouter struct {
	Router
	gate ReadyGate
}

var _ Router = &GateRouter{}

func NewGateRouter(r Router, g ReadyGate) *GateRouter {
	return &GateRouter{Router: r, gate: g}
}

func (gr *GateRouter) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	select {
	case <-gr.gate.Ready():
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return gr.Router.RouteQuery(ctx, q)
}

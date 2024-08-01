package core

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"time"
)

var _ astral.Router = &Router{}

const routingTimeout = 60 * time.Second

type Router struct {
	node *Node
	*routers.PriorityRouter
	conns sig.Map[astral.Nonce, *conn]
	pre   sig.Set[QueryFilter]
}

type QueryFilter interface {
	FilterQuery(*astral.Query) error
}

func NewRouter(node *Node) *Router {
	var router = &Router{
		node:           node,
		PriorityRouter: routers.NewPriorityRouter(),
	}

	return router
}

func (r *Router) AddPreFilter(f QueryFilter) error {
	return r.pre.Add(f)
}

func (r *Router) RouteQuery(ctx context.Context, q *astral.Query, caller io.WriteCloser) (target io.WriteCloser, err error) {
	// prefilter the query
	for _, pf := range r.pre.Clone() {
		err = pf.FilterQuery(q)
		if err != nil {
			return
		}
	}

	// log the start of routing
	if r.node.config.LogRoutingStart {
		r.node.log.Logv(2, "[%v] %v -> %v:%v routing...",
			q.Nonce, q.Caller, q.Target, q.Query,
		)
	}

	var startedAt = time.Now()
	target, err = r.routeQuery(ctx, q, caller)
	var d = time.Since(startedAt).Round(1 * time.Microsecond)

	// log routing results
	if err == nil {
		r.node.log.Infov(0, "[%v] %v -> %v:%v routed in %v",
			q.Nonce, q.Caller, q.Target, q.Query, d,
		)
	} else {
		r.node.log.Errorv(0, "[%v] %v -> %v:%v error (%v): %v",
			q.Nonce, q.Caller, q.Target, q.Query, d, err,
		)

	}

	return target, err
}

func (r *Router) routeQuery(ctx context.Context, q *astral.Query, src io.WriteCloser) (w io.WriteCloser, err error) {
	c, ok := r.conns.Set(q.Nonce, newConn(r, q))
	if !ok {
		return astral.RouteNotFound(r, errors.New("routing cycle not allowed"))
	}
	c.src = newWriter(c, src)

	ctx, cancel := context.WithTimeout(ctx, routingTimeout)
	defer cancel()

	w, err = r.PriorityRouter.RouteQuery(ctx, q, c.src)
	if err != nil {
		r.conns.Delete(q.Nonce)
		return nil, err
	}

	c.dst = newWriter(c, w)

	return c.dst, nil
}

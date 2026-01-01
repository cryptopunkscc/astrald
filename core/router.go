package core

import (
	"errors"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ astral.Router = &Router{}

const routingTimeout = 60 * time.Second

type Router struct {
	node *Node
	*routers.PriorityRouter
	conns         sig.Map[astral.Nonce, *conn]
	preprocessors sig.Set[QueryPreprocessor]
}

type QueryPreprocessor interface {
	PreprocessQuery(*QueryModifier) error
}

func NewRouter(node *Node) *Router {
	var router = &Router{
		node:           node,
		PriorityRouter: routers.NewPriorityRouter(),
	}

	return router
}

func (r *Router) AddQueryPreprocessor(f QueryPreprocessor) error {
	return r.preprocessors.Add(f)
}

func (r *Router) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (target io.WriteCloser, err error) {
	qm := &QueryModifier{query: q}
	// preprocess the query
	for _, p := range r.preprocessors.Clone() {
		err = p.PreprocessQuery(qm)
		if err != nil {
			return query.RouteNotFound(r, err)
		}

		// check if the query was blocked
		if err := qm.blocked.Get(); err != nil {
			return query.RouteNotFound(r, err)
		}
	}

	// log the start of routing
	if r.node.config.LogRoutingStart {
		r.node.log.Logv(2, "[%v] %v -> %v:%v routing...",
			&q.Nonce, q.Caller, q.Target, q.Query,
		)
	}

	var startedAt = time.Now()
	target, err = r.routeQuery(ctx, q, w)
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

func (r *Router) routeQuery(ctx *astral.Context, q *astral.Query, src io.WriteCloser) (w io.WriteCloser, err error) {
	c, ok := r.conns.Set(q.Nonce, newConn(r, q))
	if !ok {
		return query.RouteNotFound(r, errors.New("routing cycle not allowed"))
	}
	c.src = newWriter(c, src)

	actx, cancel := ctx.WithTimeout(routingTimeout)
	defer cancel()

	w, err = r.PriorityRouter.RouteQuery(actx, q, c.src)
	if err != nil {
		r.conns.Delete(q.Nonce)
		return nil, err
	}

	c.dst = newWriter(c, w)

	return c.dst, nil
}

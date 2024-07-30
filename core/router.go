package core

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"strings"
	"time"
)

var _ astral.Router = &Router{}

const routingTimeout = 60 * time.Second

type Router struct {
	*routers.PriorityRouter
	log           *log.Logger
	conns         sig.Map[astral.Nonce, *conn]
	logRouteTrace bool
}

func NewRouter(log *log.Logger) *Router {
	var router = &Router{
		PriorityRouter: routers.NewPriorityRouter(),
		log:            log,
	}

	return router
}

func (r *Router) RouteQuery(ctx context.Context, q *astral.Query, caller io.WriteCloser) (target io.WriteCloser, err error) {
	// log the start of routing
	r.log.Logv(2, "[%v] %v -> %v:%v routing...",
		q.Nonce, q.Caller, q.Target, q.Query,
	)

	var startedAt = time.Now()
	target, err = r.routeQuery(ctx, q, caller)
	var d = time.Since(startedAt).Round(1 * time.Microsecond)

	// log routing results
	if err != nil {
		r.log.Infov(0, "[%v] %v -> %v:%v error (%v): %v",
			q.Nonce, q.Caller, q.Target, q.Query, d, err,
		)

		if r.logRouteTrace {
			if rnf, ok := err.(*astral.ErrRouteNotFound); ok {
				for _, line := range strings.Split(rnf.Trace(), "\n") {
					if len(line) > 0 {
						r.log.Logv(2, "[%v] %v", q.Nonce, line)
					}
				}
			}
		}
	} else {
		r.log.Infov(0, "[%v] %v -> %v:%v routed in %v",
			q.Nonce, q.Caller, q.Target, q.Query, d,
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

func (r *Router) SetLogRouteTrace(logRouteTrace bool) {
	r.logRouteTrace = logRouteTrace
}

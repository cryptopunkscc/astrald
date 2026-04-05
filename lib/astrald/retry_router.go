package astrald

import (
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type RetryRouter struct {
	Router
	policy RetryPolicy
}

var _ Router = &RetryRouter{}

func NewRetryRouter(r Router, p RetryPolicy) *RetryRouter {
	return &RetryRouter{Router: r, policy: p}
}

func (rr *RetryRouter) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	for attempt := 0; ; attempt++ {
		conn, err := rr.Router.RouteQuery(ctx, q)
		if err == nil {
			return conn, nil
		}
		if !errors.Is(err, apphost.ErrNodeUnavailable) {
			return nil, err
		}
		d, ok := rr.policy.Next(attempt, err)
		if !ok {
			return nil, err
		}
		t := time.NewTimer(d)
		select {
		case <-t.C:
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		}
	}
}

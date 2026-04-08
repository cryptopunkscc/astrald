package astrald

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	libapphost "github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

type RetryRouter struct {
	Router
	retry       *sig.Retry
	maxAttempts int // 0 = unlimited
}

var _ Router = &RetryRouter{}

func NewRetryRouter(r Router, retry *sig.Retry) *RetryRouter {
	return &RetryRouter{Router: r, retry: retry}
}

func NewLimitedRetryRouter(r Router, retry *sig.Retry, maxAttempts int) *RetryRouter {
	return &RetryRouter{Router: r, retry: retry, maxAttempts: maxAttempts}
}

func NewNoRetryRouter(r Router) *RetryRouter {
	return &RetryRouter{Router: r, maxAttempts: 1}
}

func (rr *RetryRouter) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	for attempt := 0; ; attempt++ {
		conn, err := rr.Router.RouteQuery(ctx, q)
		if err == nil {
			if rr.retry != nil {
				rr.retry.Reset()
			}
			return conn, nil
		}
		if !errors.Is(err, libapphost.ErrNodeUnavailable) {
			return nil, err
		}
		if rr.maxAttempts > 0 && attempt+1 >= rr.maxAttempts {
			return nil, err
		}
		select {
		case <-rr.retry.Retry():
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

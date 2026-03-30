package apphost

import (
	"os"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/sig"
)

type ConnectFailurePolicy func(error) (retry bool, out error)

type RouterOption func(*Router)

type Router struct {
	endpoint             string
	token                string
	guestID              *astral.Identity
	hostID               *astral.Identity
	connectFailurePolicy ConnectFailurePolicy
}

var defaultRouter = newDefaultRouter()

func NewRouter(endpoint string, token string, opts ...RouterOption) *Router {
	r := &Router{endpoint: endpoint, token: token}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// RetryPolicy returns a ConnectFailurePolicy that retries up to maxRetries times
// using exponential backoff between minDelay and maxDelay.
func RetryPolicy(maxRetries int, minDelay, maxDelay time.Duration) ConnectFailurePolicy {
	r, _ := sig.NewRetry(minDelay, maxDelay, 2)
	return func(err error) (bool, error) {
		i := <-r.Retry()
		if i > maxRetries {
			return false, nil
		}
		return true, nil
	}
}

func WithConnectFailurePolicy(fn ConnectFailurePolicy) RouterOption {
	return func(r *Router) {
		r.connectFailurePolicy = fn
	}
}

func (r *Router) Apply(opts ...RouterOption) *Router {
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func DefaultRouter() *Router {
	return defaultRouter
}

func SetDefaultRouter(router *Router) {
	defaultRouter = router
}

// RouteQuery routes a query via the host.
func (router *Router) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	// connect to the host
	host, err := router.connect()
	if err != nil {
		return nil, err
	}

	// cancel the query with context
	var done = make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			// cancel the query on the host
			host, err := router.connect()
			if err != nil {
				return
			}
			defer host.Close()

			conn, err := host.RouteQuery(
				query.New(nil, nil, "apphost.cancel", query.Args{"id": q.Nonce}),
				astral.ZoneDevice,
				nil,
			)
			if conn != nil {
				conn.Close()
			}

			// NOTE: we're ignoring the result of the cancel op call

		case <-done:
		}
	}()

	return host.RouteQuery(q, ctx.Zone(), ctx.Filters())
}

func (router *Router) GuestID() *astral.Identity {
	if router.guestID != nil {
		return router.guestID
	}

	host, err := router.connect() // connect loads guestID
	if err != nil {
		return nil
	}
	defer host.Close()

	return router.guestID
}

func (router *Router) HostID() *astral.Identity {
	if router.hostID != nil {
		return router.hostID
	}

	host, err := router.connect() // connect loads hostID
	if err != nil {
		return nil
	}
	defer host.Close()

	return router.hostID
}

func (router *Router) Endpoint() string {
	return router.endpoint
}

func (router *Router) Protocol() string {
	split := strings.SplitN(router.endpoint, ":", 2)
	return split[0]
}

// connect establishes a new authenticated connection to the host.
func (router *Router) connect() (*Host, error) {
	for {
		host, err := Connect(router.endpoint)
		if err == nil {
			router.hostID = host.HostID()
			if len(router.token) == 0 {
				return host, nil
			}

			err = host.AuthToken(router.token)
			if err == nil {
				router.guestID = host.GuestID()
				return host, nil
			}

			host.Close()
		}

		if router.connectFailurePolicy == nil {
			return nil, err
		}

		retry, out := router.connectFailurePolicy(err)
		if retry {
			continue
		}
		if out != nil {
			return nil, out
		}
		return nil, err
	}
}

func newDefaultRouter() *Router {
	// TODO: find an optimal endpoint (unix, then tcp)
	return NewRouter(DefaultEndpoint, os.Getenv(AuthTokenEnv))
}

package apphost

import (
	"os"
	"strings"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/sig"
)

type Router struct {
	endpoint string
	token    string
	guestID  *astral.Identity
	hostID   *astral.Identity
	retry    *sig.Retry
}

var defaultRouter = newDefaultRouter()

func NewRouter(endpoint string, token string) *Router {
	r, _ := sig.NewRetry(250*time.Millisecond, 10*time.Second, 2)
	return &Router{endpoint: endpoint, token: token, retry: r}
}

func DefaultRouter() *Router {
	return defaultRouter
}

func SetDefaultRouter(router *Router) {
	defaultRouter = router
}

func (r *Router) SetRetry(retry *sig.Retry) *Router {
	r.retry = retry
	return r
}

// RouteQuery routes a query via the host, retrying the connection according to
// the router's retry policy (by default: exponential backoff until ctx is done).
func (router *Router) RouteQuery(ctx *astral.Context, q *astral.Query) (astral.Conn, error) {
	host, err := router.connectWithRetry(ctx)
	if err != nil {
		return nil, err
	}

	// cancel the query when ctx ends
	var done = make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			cancelHost, err := router.connect(ctx)
			if err != nil {
				return
			}
			defer cancelHost.Close()

			conn, _ := cancelHost.RouteQuery(
				query.New(nil, nil, apphost.MethodCancel, query.Args{"id": q.Nonce}),
				astral.ZoneDevice,
				nil,
			)
			if conn != nil {
				conn.Close()
			}

		case <-done:
		}
	}()

	return host.RouteQuery(q, ctx.Zone(), ctx.Filters())
}

func (router *Router) GuestID() *astral.Identity {
	if router.guestID != nil {
		return router.guestID
	}

	host, err := router.connect(astral.NewContext(nil))
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

	host, err := router.connect(astral.NewContext(nil))
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

// connectWithRetry retries connect until it succeeds or ctx is done.
func (router *Router) connectWithRetry(ctx *astral.Context) (*Host, error) {
	for {
		host, err := router.connect(ctx)
		if err == nil {
			router.retry.Reset()
			return host, nil
		}

		select {
		case <-router.retry.Retry():
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// connect makes a single attempt to connect and authenticate with the host.
func (router *Router) connect(ctx *astral.Context) (*Host, error) {
	host, err := Connect(ctx, router.endpoint)
	if err != nil {
		return nil, err
	}

	router.hostID = host.HostID()

	if len(router.token) == 0 {
		return host, nil
	}

	err = host.AuthToken(router.token)
	if err != nil {
		host.Close()
		return nil, err
	}

	router.guestID = host.GuestID()
	return host, nil
}

func newDefaultRouter() *Router {
	// TODO: find an optimal endpoint (unix, then tcp)
	return NewRouter(DefaultEndpoint, os.Getenv(AuthTokenEnv))
}

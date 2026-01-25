package apphost

import (
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Router struct {
	endpoint string
	token    string
	guestID  *astral.Identity
	hostID   *astral.Identity
}

var defaultRouter = newDefaultRouter()

func NewRouter(endpoint string, token string) *Router {
	return &Router{endpoint: endpoint, token: token}
}

func DefaultRouter() *Router {
	return defaultRouter
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

// Connect establishes a new authenticated connection to the host.
func (router *Router) connect() (host *Host, err error) {
	host, err = Connect(router.endpoint)
	if err != nil {
		return nil, err
	}

	router.hostID = host.HostID()

	if len(router.token) > 0 {
		err = host.AuthToken(router.token)
		if err != nil {
			return nil, err
		}

		router.guestID = host.GuestID()
	}

	return host, nil
}

func newDefaultRouter() *Router {
	return NewRouter(DefaultEndpoint, os.Getenv(AuthTokenEnv))
}

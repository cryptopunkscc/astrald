package apphost

import (
	"os"

	"github.com/cryptopunkscc/astrald/astral"
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

	return host.RouteQuery(q)
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

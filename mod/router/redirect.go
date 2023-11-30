package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/router"
)

// Redirect is a service that redirects a query to a different target
type Redirect struct {
	ServiceName string
	Node        node.Node
	Allow       id.Identity
	Query       net.Query
}

// NewRedirect creates a new redirection service on the node. Only `allow` can route to the service and the request
// will be translated to `query`.
func NewRedirect(ctx context.Context, query net.Query, allow id.Identity, node node.Node) (*Redirect, error) {
	var err error
	var r = &Redirect{
		Node:  node,
		Allow: allow,
		Query: query,
	}

	var randBytes = make([]byte, 16)
	rand.Read(randBytes)
	r.ServiceName = RouterServiceName + "." + hex.EncodeToString(randBytes)

	err = node.LocalRouter().AddRoute(r.ServiceName, r)

	return r, err
}

func (r *Redirect) RouteQuery(ctx context.Context, query net.Query, proxyCaller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	// the redirected query is locked to the caller and query nonce
	if !(query.Caller().IsEqual(r.Allow) && (query.Nonce() == r.Query.Nonce())) {
		return net.Reject()
	}

	defer r.Node.LocalRouter().RemoveRoute(r.ServiceName)

	finalQuery := r.Query

	// add identity transaltion
	mon, ok := proxyCaller.(*router.MonitoredWriter)
	if ok {
		next := mon.Output()
		var t = net.NewIdentityTranslation(next, finalQuery.Caller())
		mon.SetOutput(t)
		if s, ok := next.(net.SourceSetter); ok {
			s.SetSource(t)
		}
	} else {
		proxyCaller = net.NewIdentityTranslation(proxyCaller, finalQuery.Caller())
	}

	// reroute the query to its final destination
	target, err := r.Node.Router().RouteQuery(ctx, finalQuery, proxyCaller, hints.SetReroute().SetUpdate())
	if err != nil {
		return nil, err
	}

	if !target.Identity().IsEqual(r.Node.Identity()) {
		target = net.NewIdentityTranslation(target, r.Node.Identity())
	}

	return target, nil
}

package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/services"
)

// Redirect is a service that redirects a query to a different target
type Redirect struct {
	ServiceName string
	Node        node.Node
	Allow       id.Identity
	Query       net.Query
	service     *services.Service
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

	r.service, err = node.Services().Register(ctx, node.Identity(), r.ServiceName, r)

	return r, err
}

func (r *Redirect) RouteQuery(ctx context.Context, query net.Query, proxyCaller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if query.Caller().IsEqual(r.Allow) {
		defer r.service.Close()

		mon, ok := proxyCaller.(*node.MonitoredWriter)
		if ok {
			var t = NewIdentityTranslation(mon.Output(), r.Query.Caller())
			mon.SetOutput(t)
		} else {
			proxyCaller = NewIdentityTranslation(proxyCaller, r.Query.Caller())
		}

		if v, ok := hints.Value(node.MonitoredConnHint); ok && v != nil {
			if tracker, ok := v.(*node.MonitoredConn); ok {
				tracker.SetQuery(r.Query)
			}
		}

		target, err := r.Node.Router().RouteQuery(ctx, r.Query, proxyCaller, hints.SetDontMonitor().SetAllowRedirect())
		if err != nil {
			return nil, err
		}

		if !target.Identity().IsEqual(r.Node.Identity()) {
			target = NewIdentityTranslation(target, r.Node.Identity())
		}

		return target, nil
	}

	return net.Reject()
}

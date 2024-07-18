package relay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

// Redirect is a service that redirects a query to a different target
type Redirect struct {
	*Module
	ServiceName string
	Node        astral.Node
	Allow       id.Identity
	Query       astral.Query
}

// NewRedirect creates a new redirection service on the node. Only `allow` can route to the service and the request
// will be translated to `query`.
func NewRedirect(ctx context.Context, query astral.Query, allow id.Identity, mod *Module) (*Redirect, error) {
	var err error
	var r = &Redirect{
		Module: mod,
		Node:   mod.node,
		Allow:  allow,
		Query:  query,
	}

	var randBytes = make([]byte, 16)
	rand.Read(randBytes)
	r.ServiceName = relay.ServiceName + "." + hex.EncodeToString(randBytes)

	err = mod.AddRoute(r.ServiceName, r)

	return r, err
}

func (r *Redirect) RouteQuery(ctx context.Context, query astral.Query, proxyCaller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	// the redirected query is locked to the caller and query nonce
	if !(query.Caller().IsEqual(r.Allow) && (query.Nonce() == r.Query.Nonce())) {
		return astral.Reject()
	}

	defer r.RemoveRoute(r.ServiceName)

	finalQuery := r.Query

	// add identity transaltion
	mon, ok := proxyCaller.(*core.MonitoredWriter)
	if ok {
		next := mon.Output()
		var t = astral.NewIdentityTranslation(next, finalQuery.Caller())
		mon.SetOutput(t)
		if s, ok := next.(astral.SourceSetter); ok {
			s.SetSource(t)
		}
	} else {
		proxyCaller = astral.NewIdentityTranslation(proxyCaller, finalQuery.Caller())
	}

	// reroute the query to its final destination
	target, err := r.Node.Router().RouteQuery(ctx, finalQuery, proxyCaller, hints.SetReroute().SetUpdate())
	if err != nil {
		return nil, err
	}

	if !target.Identity().IsEqual(r.Node.Identity()) {
		target = astral.NewIdentityTranslation(target, r.Node.Identity())
	}

	return target, nil
}

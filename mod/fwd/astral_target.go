package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/router"
	"github.com/cryptopunkscc/astrald/net"
)

type AstralTarget struct {
	query  net.Query
	router net.Router
	label  string
}

func NewAstralTarget(query net.Query, router net.Router, label string) (*AstralTarget, error) {
	return &AstralTarget{
		query:  query,
		router: router,
		label:  label,
	}, nil
}

func (t *AstralTarget) RouteQuery(ctx context.Context, query net.Query, src net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return t.router.RouteQuery(
		ctx,
		t.query,
		router.NewIdentityTranslation(src, t.query.Caller()),
		net.DefaultHints(),
	)
}

func (t *AstralTarget) String() string {
	return "astral://" + t.label
}

package fwd

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type AstralTarget struct {
	query  *astral.Query
	router astral.Router
	label  string
}

func NewAstralTarget(query *astral.Query, router astral.Router, label string) (*AstralTarget, error) {
	return &AstralTarget{
		query:  query,
		router: router,
		label:  label,
	}, nil
}

func (t *AstralTarget) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return t.router.RouteQuery(
		ctx,
		astral.NewQuery(t.query.Caller, t.query.Target, t.query.Query),
		w,
	)
}

func (t *AstralTarget) String() string {
	return "astral://" + t.label
}

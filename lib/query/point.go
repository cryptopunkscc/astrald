package query

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

type Point struct {
	router   astral.Router
	sourceID *astral.Identity
	targetID *astral.Identity
}

func NewPoint(router astral.Router, sourceID *astral.Identity, targetID *astral.Identity) *Point {
	return &Point{router: router, sourceID: sourceID, targetID: targetID}
}

func (t *Point) Query(ctx context.Context, method string, args any) (astral.Conn, error) {
	return Route(ctx, t.router, New(t.sourceID, t.targetID, method, args))
}

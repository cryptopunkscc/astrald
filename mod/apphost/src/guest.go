package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"io"
	"sync/atomic"
)

type Guest struct {
	Identity *astral.Identity

	router *routers.PrefixRouter
	count  atomic.Int32
}

func (guest *Guest) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
	return guest.router.RouteQuery(ctx, query, caller)
}

func NewGuest(identity *astral.Identity) *Guest {
	return &Guest{
		Identity: identity,
		router:   routers.NewPrefixRouter(false),
	}
}

func (guest *Guest) AddRoute(name string, target astral.Router) error {
	if err := guest.router.AddRoute(name, target); err != nil {
		return err
	}
	guest.count.Add(1)

	return nil
}

func (guest *Guest) RemoveRoute(name string) error {
	if err := guest.router.RemoveRoute(name); err != nil {
		return err
	}
	guest.count.Add(-1)

	return nil
}

func (guest *Guest) RouteCount() int {
	return int(guest.count.Load())
}

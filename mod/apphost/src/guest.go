package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/net"
	"sync/atomic"
)

type Guest struct {
	Identity id.Identity

	router *routers.PrefixRouter
	count  atomic.Int32
}

func (guest *Guest) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return guest.router.RouteQuery(ctx, query, caller, hints)
}

func NewGuest(identity id.Identity) *Guest {
	return &Guest{
		Identity: identity,
		router:   routers.NewPrefixRouter(false),
	}
}

func (guest *Guest) AddRoute(name string, target net.Router) error {
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

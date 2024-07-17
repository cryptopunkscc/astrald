package routers

import (
	"cmp"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ net.Router = &PriorityRouter{}

type PriorityRouter struct {
	entries sig.Set[*Entry]
}

type PriorityAdder interface {
	Add(r net.Router, prio int) error
}

var _ PriorityAdder = &PriorityRouter{}

type Entry struct {
	Router net.Router
	Prio   int
}

func NewPriorityRouter() *PriorityRouter {
	return &PriorityRouter{}
}

func (router *PriorityRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (w net.SecureWriteCloser, err error) {
	var errs []error

	for _, r := range router.entries.Clone() {
		w, err = r.Router.RouteQuery(ctx, query, caller, hints)
		switch {
		case err == nil:
			return
		case errors.Is(err, net.ErrRejected):
			return
		default:
			errs = append(errs, err)
		}
	}

	return net.RouteNotFound(router, errs...)
}

func (router *PriorityRouter) Add(r net.Router, prio int) error {
	router.entries.Add(&Entry{Router: r, Prio: prio})

	router.entries.Sort(func(a, b *Entry) int {
		return cmp.Compare(a.Prio, b.Prio)
	})

	return nil
}

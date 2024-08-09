package routers

import (
	"cmp"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
)

var _ astral.Router = &PriorityRouter{}

type PriorityRouter struct {
	entries sig.Set[*Entry]
}

type PriorityAdder interface {
	Add(r astral.Router, prio int) error
}

var _ PriorityAdder = &PriorityRouter{}

type Entry struct {
	Router astral.Router
	Prio   int
}

func NewPriorityRouter() *PriorityRouter {
	return &PriorityRouter{}
}

func (router *PriorityRouter) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (rw io.WriteCloser, err error) {
	var errs []error

	for _, r := range router.entries.Clone() {
		rw, err = r.Router.RouteQuery(ctx, q, w)
		switch {
		case err == nil:
			return
		case errors.Is(err, &astral.ErrRejected{}):
			return
		default:
			errs = append(errs, err)
		}
	}

	return astral.RouteNotFound(router, errs...)
}

func (router *PriorityRouter) Add(r astral.Router, prio int) error {
	router.entries.Add(&Entry{Router: r, Prio: prio})

	router.entries.Sort(func(a, b *Entry) int {
		return cmp.Compare(a.Prio, b.Prio)
	})

	return nil
}

package dir

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opApplyFiltersArgs struct {
	Filters string
	ID      string `query:"optional"`
	In      string `query:"optional"`
	Out     string `query:"optional"`
}

func (mod *Module) OpApplyFilters(ctx *astral.Context, q *routing.IncomingQuery, args opApplyFiltersArgs) (err error) {
	ch := q.Accept(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// set initial values
	var (
		identity = q.Caller()
		filters  = strings.Split(args.Filters, ",")
	)

	// parse arg
	if len(args.ID) > 0 {
		identity, err = mod.ResolveIdentity(args.ID)
		if err != nil {
			return ch.Send(astral.Err(err))
		}
	}

	res := mod.ApplyFilters(identity, filters...)

	return ch.Send((*astral.Bool)(&res))
}

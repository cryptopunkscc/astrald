package dir

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opApplyFiltersArgs struct {
	Arg string
	ID  string `query:"optional"`
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpApplyFilters(ctx *astral.Context, q *ops.Query, args opApplyFiltersArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// set initial values
	var (
		identity = q.Caller()
		filters  = strings.Split(args.Arg, ",")
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

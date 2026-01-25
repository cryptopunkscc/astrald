package dir

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opFiltersArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpFilters(ctx *astral.Context, q shell.Query, args opFiltersArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	for _, f := range mod.Filters() {
		err = ch.Send(astral.NewString8(f))
		if err != nil {
			return
		}
	}

	return ch.Send(&astral.EOS{})
}

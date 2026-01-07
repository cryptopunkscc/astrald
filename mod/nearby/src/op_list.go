package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpList(ctx *astral.Context, q shell.Query, args opListArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for _, v := range mod.Cache().Clone() {
		err = ch.Send(&nearby.Status{
			Identity:    v.Identity,
			Alias:       v.Status.Alias,
			Attachments: v.Status.Attachments,
		})
		if err != nil {
			return
		}
	}

	return
}

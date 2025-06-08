package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nearby"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opListArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpList(ctx *astral.Context, q shell.Query, args opListArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for _, v := range mod.Cache().Clone() {
		err = ch.Write(&nearby.Status{
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

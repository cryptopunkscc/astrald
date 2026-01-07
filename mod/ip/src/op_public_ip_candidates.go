package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opPublicIPCandidatesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpPublicIPCandidates(ctx *astral.Context, q shell.Query, args opPublicIPCandidatesArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for _, addr := range mod.PublicIPCandidates() {
		err = ch.Send(&addr)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})

}

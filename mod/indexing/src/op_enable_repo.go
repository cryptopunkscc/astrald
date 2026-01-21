package indexing

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opEnableRepoArgs struct {
	Repo    string
	Disable bool   `query:"optional"`
	In      string `query:"optional"`
	Out     string `query:"optional"`
}

func (mod *Module) OpEnableRepo(ctx *astral.Context, q shell.Query, args opEnableRepoArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	repo := mod.Objects.GetRepository(args.Repo)
	if repo == nil {
		return ch.Send(astral.NewError("repository not found"))
	}

	if args.Disable {
		err = mod.DisableRepo(ctx, args.Repo)
	} else {
		err = mod.EnableRepo(ctx, args.Repo)
	}

	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}

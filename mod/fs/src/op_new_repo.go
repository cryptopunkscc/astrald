package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewRepoArgs struct {
	Path  string
	Label string
	In    string `query:"optional"`
	Out   string `query:"optional"`
}

func (mod *Module) OpNewRepo(ctx *astral.Context, q shell.Query, args opNewRepoArgs) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var repo objects.Repository

	repo = NewRepository(mod, args.Label, args.Path)

	err = mod.Objects.AddRepository(args.Label, repo)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = mod.Objects.AddGroup(objects.RepoLocal, args.Label)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}

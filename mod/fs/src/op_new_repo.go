package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opNewRepoArgs struct {
	Path  string
	Name  string
	Label string `query:"optional"`
	In    string `query:"optional"`
	Out   string `query:"optional"`
}

func (mod *Module) OpNewRepo(ctx *astral.Context, q *ops.Query, args opNewRepoArgs) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Label == "" {
		args.Label = args.Name
	}

	var repo objects.Repository

	repo = NewRepository(mod, args.Name, args.Path)

	err = mod.Objects.AddRepository(args.Name, repo)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	err = mod.Objects.AddGroup(objects.RepoLocal, args.Name)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}

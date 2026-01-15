package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRemoveRepo struct {
	Repo string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpRemoveRepo(ctx *astral.Context, q shell.Query, args opRemoveRepo) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	repo := mod.Objects.GetRepository(args.Repo)
	if repo == nil {
		return ch.Send(astral.NewError("repo not found"))
	}

	// If it's a WatchRepository, stop background activity and cleanup DB entries
	if wr, ok := repo.(*WatchRepository); ok {
		_ = wr.Close()
		wr.Cleanup()
	}

	err = mod.Objects.RemoveRepository(args.Repo)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = mod.Objects.RemoveGroup(objects.RepoLocal, args.Repo)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}

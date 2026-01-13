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

type repoCloser interface {
	Close() error
}

func (mod *Module) OpRemoveRepo(ctx *astral.Context, q shell.Query, args opRemoveRepo) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	repo := mod.Objects.GetRepository(args.Repo)
	if repo == nil {
		return ch.Send(astral.NewError("repo not found"))
	}

	// Stop repository background activity (watchers / scans), if supported.
	if c, ok := repo.(repoCloser); ok {
		_ = c.Close()
	}

	// Drop queued/rerun indexing work for this repo label.
	mod.pathIndexer.DropOwner(repo.Label())

	// Unregister from repository group and from objects module.
	_ = mod.Objects.RemoveGroup(objects.RepoLocal, args.Repo)
	if err := mod.Objects.RemoveRepository(args.Repo); err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	// Keep local mirror map consistent if it was used.
	_, _ = mod.repos.Delete(args.Repo)

	return ch.Send(&astral.Ack{})
}

package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRepositoriesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpRepositories(ctx *astral.Context, q shell.Query, args opRepositoriesArgs) (err error) {
	ctx = ctx.ExcludeZone(astral.ZoneNetwork)

	ch := q.AcceptChannel(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for name, repo := range mod.repos.Clone() {
		free, _ := repo.Free(ctx)

		err = ch.Send(&objects.RepositoryInfo{
			Name:  astral.String8(name),
			Label: astral.String8(repo.Label()),
			Free:  astral.Int64(free),
		})
		if err != nil {
			return
		}
	}

	return ch.Send(&astral.EOS{})
}

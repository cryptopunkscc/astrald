package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRepositoriesArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpRepositories(ctx *astral.Context, q shell.Query, args opRepositoriesArgs) (err error) {
	ctx = ctx.ExcludeZone(astral.ZoneNetwork)

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()
	
	for id, repo := range mod.repos.Clone() {
		free, _ := repo.Free(ctx)

		err = ch.Write(&objects.RepositoryInfo{
			ID:    astral.String8(id),
			Label: astral.String8(repo.Label()),
			Free:  astral.Uint64(free),
		})
		if err != nil {
			return
		}
	}

	return
}

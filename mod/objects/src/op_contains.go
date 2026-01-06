package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opContainsArgs struct {
	ID   *astral.ObjectID
	Out  string        `query:"optional"`
	Zone astral.Zone   `query:"optional"`
	Repo astral.String `query:"optional"`
}

func (mod *Module) OpContains(ctx *astral.Context, q shell.Query, args opContainsArgs) (err error) {
	ctx = ctx.WithIdentity(q.Caller()).WithZone(args.Zone)

	repo := mod.GetRepository(args.Repo.String())
	if repo == nil {
		return q.RejectWithCode(astral.CodeInvalidQuery)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	has, err := repo.Contains(ctx, args.ID)

	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write((*astral.Bool)(&has))
}

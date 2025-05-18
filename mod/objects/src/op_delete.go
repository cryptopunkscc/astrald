package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opDeleteArgs struct {
	ID   *astral.ObjectID
	Out  string `query:"optional"`
	Repo string `query:"optional"`
}

func (mod *Module) OpDelete(ctx *astral.Context, q shell.Query, args opDeleteArgs) (err error) {
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.RejectWithCode(8)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	err = repo.Delete(ctx.WithIdentity(q.Caller()), args.ID)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	return ch.Write(&astral.Ack{})
}

package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opDeleteArgs struct {
	ID     *object.ID
	Format string `query:"optional"`
	Repo   string `query:"optional"`
}

func (mod *Module) OpDelete(ctx *astral.Context, q shell.Query, args opDeleteArgs) (err error) {
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	err = repo.Delete(ctx.WithIdentity(q.Caller()), args.ID)
	if err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}
	return ch.Write(&astral.Ack{})
}

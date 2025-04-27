package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type opDeleteAllArgs struct {
	Format string `query:"optional"`
	From   string `query:"optional"`
	Repo   string `query:"optional"`
}

func (mod *Module) OpDeleteAll(ctx *astral.Context, q shell.Query, args opDeleteAllArgs) (err error) {
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.Reject()
	}

	chOut := astral.NewChannel(q.Accept(), args.Format)
	chIn := astral.NewChannel(chOut.Transport(), args.From)
	defer chOut.Close()

	for {
		o, err := chIn.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		objectID, ok := o.(*object.ID)
		if !ok {
			chOut.Write(astral.NewError("not an object ID"))
			continue
		}

		err = repo.Delete(ctx.WithIdentity(q.Caller()), objectID)
		if err != nil {
			chOut.Write(astral.NewError(err.Error()))
			continue
		}

		chOut.Write(&astral.Ack{})
	}
}

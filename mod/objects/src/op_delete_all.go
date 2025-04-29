package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type opDeleteAllArgs struct {
	Out  string `query:"optional"`
	In   string `query:"optional"`
	Repo string `query:"optional"`
}

func (mod *Module) OpDeleteAll(ctx *astral.Context, q shell.Query, args opDeleteAllArgs) (err error) {
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	for {
		o, err := ch.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		objectID, ok := o.(*object.ID)
		if !ok {
			err = ch.Write(astral.NewError("not an object ID"))
			if err != nil {
				return err
			}
			continue
		}

		err = repo.Delete(ctx.WithIdentity(q.Caller()), objectID)
		if err != nil {
			err = ch.Write(astral.NewError(err.Error()))
			if err != nil {
				return err
			}
			continue
		}

		err = ch.Write(&astral.Ack{})
		if err != nil {
			return err
		}
	}
}

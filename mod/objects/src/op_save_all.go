package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opSaveAllArgs struct {
	From   string `query:"optional"`
	Format string `query:"optional"`
}

func (mod *Module) OpSaveAll(ctx *astral.Context, q shell.Query, args opSaveAllArgs) (err error) {
	chOut := astral.NewChannel(q.Accept(), args.Format)
	chIn := astral.NewChannel(chOut.Transport(), args.From)
	defer chOut.Close()

	for {
		object, err := chIn.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		objectID, err := objects.Save(ctx, object, mod.Root())
		if err != nil {
			err = chOut.Write(astral.NewError(err.Error()))
			if err != nil {
				return err
			}
			continue
		}

		err = chOut.Write(objectID)
		if err != nil {
			return err
		}
	}
}

package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opSaveAllArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSaveAll(ctx *astral.Context, q shell.Query, args opSaveAllArgs) (err error) {
	ch := astral.NewChannelAsym(q.Accept(), args.In, args.Out)
	defer ch.Close()

	for {
		object, err := ch.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		objectID, err := objects.Save(ctx, object, mod.Root())
		if err != nil {
			err = ch.Write(astral.NewError(err.Error()))
			if err != nil {
				return err
			}
			continue
		}

		err = ch.Write(objectID)
		if err != nil {
			return err
		}
	}
}

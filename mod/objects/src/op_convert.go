package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opConvertArgs struct {
	From   string `query:"optional"`
	Format string `query:"optional"`
}

func (mod *Module) OpConvert(ctx *astral.Context, q shell.Query, args opConvertArgs) (err error) {
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

		err = chOut.Write(object)
		if err != nil {
			return err
		}
	}
}

package objects

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opConvertArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpConvert(ctx *astral.Context, q shell.Query, args opConvertArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	for {
		object, err := ch.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		err = ch.Write(object)
		if err != nil {
			return err
		}
	}
}

package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

const defaultMemSize = 64 * astral.MiB

type opNewMemArgs struct {
	Name string
	Size string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpNewMem(ctx *astral.Context, q shell.Query, args opNewMemArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	size := defaultMemSize
	if len(args.Size) > 0 {
		size, err = astral.ParseSize(args.Size)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	err = mod.NewMem(args.Name, uint64(size))
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}

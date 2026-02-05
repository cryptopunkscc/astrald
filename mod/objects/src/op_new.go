package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opNewArgs struct {
	Type string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpNew(ctx *astral.Context, query *ops.Query, args opNewArgs) (err error) {
	ch := query.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	object := astral.New(args.Type)
	if object == nil {
		return ch.Send(&astral.Nil{})
	}

	return ch.Send(object)
}

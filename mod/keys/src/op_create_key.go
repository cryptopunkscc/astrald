package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opCreateKeyArgs struct {
	Alias string
	Out   string `query:"optional"`
}

func (mod *Module) OpCreateKey(_ *astral.Context, q *ops.Query, args opCreateKeyArgs) (err error) {
	mod.log.Infov(1, "creating key for %v", args.Alias)

	key, _, err := mod.CreateKey(args.Alias)
	if err != nil {
		mod.log.Errorv(1, "error creating key for %v: %v", args.Alias, err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(key)
}

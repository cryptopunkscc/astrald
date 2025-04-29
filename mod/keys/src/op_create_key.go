package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opCreateKeyArgs struct {
	Alias string
	Out   string `query:"optional"`
}

func (mod *Module) OpCreateKey(_ *astral.Context, q shell.Query, args opCreateKeyArgs) (err error) {
	mod.log.Infov(1, "creating key for %v", args.Alias)

	key, _, err := mod.CreateKey(args.Alias)
	if err != nil {
		mod.log.Errorv(1, "error creating key for %v: %v", args.Alias, err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(key)
}

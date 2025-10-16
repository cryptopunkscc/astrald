package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opLocalAddrsArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpLocalAddrs(ctx *astral.Context, q shell.Query, args opLocalAddrsArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	addrs, err := mod.localAddresses(false)
	if err != nil {
		return
	}

	for _, addr := range addrs {
		err = ch.Write(&addr)
		if err != nil {
			return
		}
	}

	return
}

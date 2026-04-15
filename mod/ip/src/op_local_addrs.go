package ip

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opLocalAddrsArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpLocalAddrs(ctx *astral.Context, q *routing.IncomingQuery, args opLocalAddrsArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	addrs, err := mod.localAddresses(false)
	if err != nil {
		return
	}

	for _, addr := range addrs {
		err = ch.Send(&addr)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}

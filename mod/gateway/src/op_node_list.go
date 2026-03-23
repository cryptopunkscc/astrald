package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type opListArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpNodeList(ctx *astral.Context, q *ops.Query, args opListArgs) error {
	ch := q.AcceptChannel(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	for _, client := range mod.registeredNodes.Values() {
		if client.GetVisibility() != gateway.VisibilityPublic {
			continue
		}
		if err := ch.Send(client.Identity); err != nil {
			return err
		}
	}

	return ch.Send(&astral.EOS{})
}

package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opRelayArgs struct {
	Nonce     astral.Nonce
	SetCaller *astral.Identity `query:"optional"`
	SetTarget *astral.Identity `query:"optional"`
}

func (mod *Module) OpRelay(ctx *astral.Context, q shell.Query, args opRelayArgs) (err error) {
	ch := astral.NewChannel(q.Accept())
	defer ch.Close()

	r, _ := mod.relays.Set(args.Nonce, &Relay{})

	if !args.SetCaller.IsZero() {
		if !args.SetCaller.IsEqual(q.Caller()) {
			if !mod.Auth.Authorize(q.Caller(), nodes.ActionRelayFor, args.SetCaller) {
				return ch.Write(astral.NewError("unauthorized"))
			}
		}
		r.Caller = args.SetCaller
	}

	if !args.SetTarget.IsZero() {
		r.Target = args.SetTarget
	}

	return ch.Write(&astral.Ack{})
}

package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type opBindArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpBind(ctx *astral.Context, q *routing.IncomingQuery, args opBindArgs) error {
	// only local apps can bind
	if q.Origin() == astral.OriginNetwork {
		return q.Reject()
	}

	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if err := ch.Send(&astral.Ack{}); err != nil {
		return err
	}

	var actions []func() error
	defer func() {
		for _, a := range actions {
			if err := a(); err != nil {
				mod.log.Error("bind error: %v", err)
			}
		}
	}()

	return ch.Switch(
		channel.WithContext(ctx),
		func(msg *apphost.BindMsg) error {
			actions = append(actions, func() error {
				return mod.removeHandlersByToken(msg.Token)
			})
			return nil
		},
	)
}

package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type opNodePunchArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpNodePunch(ctx *astral.Context, q *routing.IncomingQuery, args opNodePunchArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	localIP, err := mod.getLocalIPv4()
	if err != nil {
		return err
	}

	mod.log.Log("starting traversal as participant with %v", q.Caller())
	proto := nat.NewPunchProtocol(ctx.Identity(), q.Caller(), localIP)

	err = ch.Switch(
		proto.ExpectSignal(nat.PunchSignalTypeOffer, proto.OnOffer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	puncher, err := mod.newPuncher(proto.Session)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			puncher.Close()
		}
	}()

	localPort, err := puncher.Open()
	if err != nil {
		return err
	}
	proto.LocalPort = astral.Uint16(localPort)

	if err = ch.Send(proto.AnswerSignal()); err != nil {
		return err
	}

	err = ch.Switch(
		proto.ExpectSignal(nat.PunchSignalTypeReady, nil),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err = ch.Send(proto.GoSignal()); err != nil {
		return err
	}

	result, err := puncher.HolePunch(ctx, proto.PeerIP, int(proto.PeerPort))
	if err != nil {
		return err
	}

	proto.SetPunchResult(result)

	err = ch.Switch(
		proto.ExpectSignal(nat.PunchSignalTypeResult, proto.OnResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err = ch.Send(proto.ResultSignal()); err != nil {
		return err
	}

	puncher.Close()

	mod.log.Info("NAT traversal succeeded with %v: %v <-> %v", q.Caller(), proto.Hole.ActiveEndpoint, proto.Hole.PassiveEndpoint)

	mod.addHole(proto.Hole, false)
	return nil
}

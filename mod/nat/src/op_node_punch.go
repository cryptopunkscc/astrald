package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type opNodePunchArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpNodePunch(ctx *astral.Context, q *ops.Query, args opNodePunchArgs) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	localIP, err := mod.getLocalIPv4()
	if err != nil {
		return err
	}

	mod.log.Log("starting traversal as participant with %v", q.Caller())
	traversal := nat.NewTraversal(ctx.Identity(), q.Caller(), localIP)

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeOffer, traversal.OnOffer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	puncher, err := mod.newPuncher(traversal.Session)
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
	traversal.LocalPort = astral.Uint16(localPort)

	if err = ch.Send(traversal.AnswerSignal()); err != nil {
		return err
	}

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeReady, nil),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err = ch.Send(traversal.GoSignal()); err != nil {
		return err
	}

	result, err := puncher.HolePunch(ctx, traversal.PeerIP, int(traversal.PeerPort))
	if err != nil {
		return err
	}

	traversal.SetPunchResult(result)

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeResult, traversal.OnResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	if err = ch.Send(traversal.ResultSignal()); err != nil {
		return err
	}

	puncher.Close()

	mod.log.Info("NAT traversal succeeded with %v: %v <-> %v", q.Caller(), traversal.Hole.ActiveEndpoint, traversal.Hole.PassiveEndpoint)

	mod.addHole(traversal.Hole, false)
	return nil
}

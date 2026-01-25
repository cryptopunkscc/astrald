package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opMigrateSessionArgs struct {
	SessionID astral.Nonce
	StreamID  astral.Nonce
	Start     astral.Bool `query:"optional"`
	Out       string      `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q *ops.Query, args opMigrateSessionArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if args.SessionID == 0 {
		return ch.Send(astral.NewError("missing sessionId"))
	}
	if args.StreamID == 0 {
		return ch.Send(astral.NewError("missing stream ids"))
	}

	isInitiator := args.Start
	if isInitiator {
		sessionToMigrate, ok := mod.peers.sessions.Get(args.SessionID)
		if !ok {
			return ch.Send(astral.NewError("session not found"))
		}

		target := sessionToMigrate.RemoteIdentity
		args := &opMigrateSessionArgs{
			SessionID: args.SessionID,
			StreamID:  args.StreamID,
			Start:     false,
			Out:       args.Out,
		}

		// Route a channel to the remote OpMigrateSession
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, nodes.MethodMigrateSession, &args),
		)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		ms, err := mod.createSessionMigrator(ctx, RoleInitiator, peerCh, target, args.SessionID, args.StreamID)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
		return ch.Send(&astral.Ack{})
	}

	ms, err := mod.createSessionMigrator(ctx, RoleResponder, ch, q.Caller(),
		args.SessionID, args.StreamID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	err = ms.Run(ctx)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}

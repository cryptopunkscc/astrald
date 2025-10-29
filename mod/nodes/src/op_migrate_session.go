package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	modnodes "github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMigrateSessionArgs struct {
	SessionID astral.Nonce
	StreamID  astral.Nonce
	Start     astral.Bool `query:"optional"`
	Out       string      `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q shell.Query, args opMigrateSessionArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	if args.SessionID == 0 {
		return ch.Write(astral.NewError("missing sessionId"))
	}
	if args.StreamID == 0 {
		return ch.Write(astral.NewError("missing stream ids"))
	}

	isInitiator := args.Start
	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.SessionID)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity

		// Route a channel to the remote OpMigrateSession
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, modnodes.MethodMigrateSession, &opMigrateSessionArgs{
				SessionID: args.SessionID,
				StreamID:  args.StreamID,
				Start:     false,
				Out:       args.Out,
			}),
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		ms := &sessionMigrator{
			mod:       mod,
			sess:      sess,
			role:      RoleInitiator,
			ch:        peerCh,
			local:     ctx.Identity(),
			peer:      target,
			sessionId: args.SessionID,
			streamId:  args.StreamID,
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		return ch.Write(&astral.Ack{})
	}

	// Responder: attach local session and pass local streamId to FSM
	sess, ok := mod.peers.sessions.Get(args.SessionID)
	if !ok || sess == nil {
		return ch.Write(astral.NewError("session not found"))
	}

	ms := &sessionMigrator{
		mod:       mod,
		sess:      sess,
		role:      RoleResponder,
		ch:        ch,
		local:     ctx.Identity(),
		peer:      q.Caller(),
		sessionId: args.SessionID,
		streamId:  args.StreamID,
	}

	if err := ms.Run(ctx); err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}

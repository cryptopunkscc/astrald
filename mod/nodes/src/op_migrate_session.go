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

	mod.log.Log("OpMigrateSession start initiator %v session %v stream %v", q.Origin() != astral.OriginNetwork, args.SessionID, args.StreamID)

	if args.SessionID == 0 {
		return ch.Write(astral.NewError("missing sessionId"))
	}
	if args.StreamID == 0 {
		return ch.Write(astral.NewError("missing stream ids"))
	}

	isInitiator := q.Origin() != astral.OriginNetwork

	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.SessionID)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			mod.log.Log("OpMigrateSession session not found for %v %v", args.SessionID, args.StreamID)
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity
		mod.log.Log("OpMigrateSession initiator routing to %v %v", target, args.SessionID)

		// Route a channel to the remote OpMigrateSession with peer's local stream id in StreamID
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, modnodes.MethodMigrateSession, &opMigrateSessionArgs{
				SessionID: args.SessionID,
				StreamID:  args.StreamID, // remote's local id
				Start:     false,
				Out:       args.Out,
			}),
		)
		if err != nil {
			mod.log.Log("OpMigrateSession initiator route error %v %v", args.SessionID, err)
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()
		mod.log.Log("OpMigrateSession initiator control channel ready %v %v", target, args.SessionID)

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
			mod.log.Log("OpMigrateSession initiator FSM error %v %v", args.SessionID, err)
			return ch.Write(astral.NewError(err.Error()))
		}
		mod.log.Log("OpMigrateSession initiator completed %v %v", args.SessionID, args.StreamID)
		return ch.Write(&astral.Ack{})
	}

	// Responder: attach local session and pass local streamId to FSM
	sess, ok := mod.peers.sessions.Get(args.SessionID)
	if !ok || sess == nil {
		mod.log.Log("OpMigrateSession responder session not found %v %v", args.SessionID, args.StreamID)
		return ch.Write(astral.NewError("session not found"))
	}
	mod.log.Log("OpMigrateSession responder setup caller %v %v", q.Caller(), args.SessionID)

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
		mod.log.Log("OpMigrateSession responder FSM error %v %v", args.SessionID, err)
		return ch.Write(astral.NewError(err.Error()))
	}

	mod.log.Log("OpMigrateSession responder completed %v %v", args.SessionID, args.StreamID)
	return ch.Write(&astral.Ack{})
}

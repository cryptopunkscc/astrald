package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	modnodes "github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMigrateSessionArgs struct {
	Nonce        astral.Nonce
	Start        astral.Bool `query:"optional"`
	StreamID     astral.Int64
	PeerStreamID astral.Int64
	Out          string `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q shell.Query, args opMigrateSessionArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	if args.Nonce == 0 {
		return ch.Write(astral.NewError("missing nonce"))
	}
	if args.StreamID == 0 || args.PeerStreamID == 0 {
		return ch.Write(astral.NewError("missing stream ids"))
	}

	isInitiator := q.Origin() != astral.OriginNetwork

	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.Nonce)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity

		// Route a channel to the remote OpMigrateSession with both IDs swapped for the peer
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, modnodes.MethodMigrateSession, &opMigrateSessionArgs{
				Out:          args.Out,
				Nonce:        args.Nonce,
				StreamID:     args.PeerStreamID, // peer's local id
				PeerStreamID: args.StreamID,     // our local id as their peer id
			}),
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		ms := &sessionMigrator{
			mod:           mod,
			sess:          sess,
			role:          RoleInitiator,
			ch:            peerCh,
			local:         ctx.Identity(),
			peer:          target,
			nonce:         args.Nonce,
			localTargetID: args.StreamID,
			peerTargetID:  args.PeerStreamID,
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		return ch.Write(&astral.Ack{})
	}

	// Responder: both IDs must be present; attach local session and pass IDs to FSM
	sess, ok := mod.peers.sessions.Get(args.Nonce)
	if !ok || sess == nil {
		return ch.Write(astral.NewError("session not found"))
	}

	ms := &sessionMigrator{
		mod:           mod,
		sess:          sess,
		role:          RoleResponder,
		ch:            ch,
		local:         ctx.Identity(),
		peer:          q.Caller(),
		nonce:         args.Nonce,
		localTargetID: args.StreamID,
		peerTargetID:  args.PeerStreamID,
	}

	if err := ms.Run(ctx); err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}

package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	modnodes "github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMigrateSessionArgs struct {
	// Nonce identifies the session for migration signaling.
	Nonce astral.Nonce
	// Out selects the output format for this op.
	Out string `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q shell.Query, args opMigrateSessionArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	if args.Nonce == 0 {
		return ch.Write(astral.NewError("missing nonce"))
	}

	// Initiator if this call did not come from the network; otherwise responder.
	isInitiator := q.Origin() != astral.OriginNetwork

	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.Nonce)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity

		// Route a channel to the remote OpMigrateSession, passing only the nonce.
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, modnodes.MethodMigrateSession, &opMigrateSessionArgs{
				Out:   args.Out,
				Nonce: args.Nonce,
			}),
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		// Capture local stream id if available
		var localStreamID astral.Int64
		if sess.stream != nil {
			localStreamID = astral.Int64(sess.stream.id)
		}

		ms := &sessionMigrator{
			mod:   mod,
			role:  RoleInitiator,
			ch:    peerCh,
			local: ctx.Identity(),
			peer:  target,
			nonce: args.Nonce,
			link:  modnodes.LinkSelector{Identity: target, StreamId: localStreamID},
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		return ch.Write(&astral.Ack{})
	}

	// Responder branch: use accepted channel directly.
	ms := &sessionMigrator{
		mod:   mod,
		role:  RoleResponder,
		ch:    ch,
		local: ctx.Identity(),
		peer:  q.Caller(),
		nonce: args.Nonce,
	}

	if err := ms.Run(ctx); err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}

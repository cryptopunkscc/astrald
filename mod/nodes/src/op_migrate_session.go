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
	// Start indicates if this node is the initiator of the migration.
	Start astral.Bool `query:"optional"`
	// StreamID is this node's target stream id (local, optional).
	StreamID astral.Int64 `query:"optional"`
	// PeerStreamID is the last known peer's target stream id (optional, for tracing).
	PeerStreamID astral.Int64 `query:"optional"`
	// Out selects the output format for this op.
	Out string `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q shell.Query, args opMigrateSessionArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.Out, args.Out)
	defer ch.Close()

	if args.Nonce == 0 {
		return ch.Write(astral.NewError("missing nonce"))
	}

	isInitiator := q.Origin() != astral.OriginNetwork

	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.Nonce)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity

		if args.StreamID == 0 {
			return ch.Write(astral.NewError("missing stream_id"))
		}
		tgt := mod.findStreamByID(int(args.StreamID))
		if tgt == nil {
			return ch.Write(astral.NewError("target stream not found"))
		}

		// Prepare session to migrate to the selected target stream
		if err := sess.Migrate(tgt); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}

		// Route a channel to the remote OpMigrateSession, passing only the nonce.
		peerCh, err := query.RouteChan(
			ctx.IncludeZone(astral.ZoneNetwork),
			mod.node,
			query.New(ctx.Identity(), target, modnodes.MethodMigrateSession, &opMigrateSessionArgs{
				Out:          args.Out,
				Nonce:        args.Nonce,
				StreamID:     args.PeerStreamID, // pass peer's local target id, if provided
				PeerStreamID: args.StreamID,
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
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		return ch.Write(&astral.Ack{})
	}

	// Responder: must receive local StreamID (its own) or error
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
	}

	if err := ms.Run(ctx); err != nil {
		return ch.Write(astral.NewError(err.Error()))
	}

	return ch.Write(&astral.Ack{})
}

// findStreamByID scans local streams and returns the stream with given id.
func (mod *Module) findStreamByID(id int) *Stream {
	for _, s := range mod.peers.streams.Clone() {
		if s.id == id {
			return s
		}
	}
	return nil
}

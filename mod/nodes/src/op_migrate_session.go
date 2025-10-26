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
	Start bool `query:"optional"`
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

	isInitiator := args.Start

	if isInitiator {
		// Find the session locally to determine the remote identity.
		sess, ok := mod.peers.sessions.Get(args.Nonce)
		if !ok || sess == nil || sess.RemoteIdentity == nil || sess.RemoteIdentity.IsZero() {
			return ch.Write(astral.NewError("session not found"))
		}
		target := sess.RemoteIdentity

		// Resolve local target stream
		var tgt *Stream
		if args.StreamID != 0 {
			tgt = mod.findStreamByID(int(args.StreamID))
		} else {
			tgt = mod.pickAltStream(sess.stream)
		}
		if tgt == nil {
			return ch.Write(astral.NewError("no target stream available"))
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
				Out:   args.Out,
				Nonce: args.Nonce,
			}),
		)
		if err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		defer peerCh.Close()

		ms := &sessionMigrator{
			mod:   mod,
			sess:  sess,
			role:  RoleInitiator,
			ch:    peerCh,
			local: ctx.Identity(),
			peer:  target,
			nonce: args.Nonce,
			link:  modnodes.LinkSelector{Identity: target, StreamId: astral.Int64(tgt.id)},
		}

		if err := ms.Run(ctx); err != nil {
			return ch.Write(astral.NewError(err.Error()))
		}
		return ch.Write(&astral.Ack{})
	}

	// Responder branch: use accepted channel directly and attach local session
	sess, ok := mod.peers.sessions.Get(args.Nonce)
	if !ok || sess == nil {
		return ch.Write(astral.NewError("session not found"))
	}

	ms := &sessionMigrator{
		mod:   mod,
		sess:  sess,
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

// findStreamByID scans local streams and returns the stream with given id.
func (mod *Module) findStreamByID(id int) *Stream {
	for _, s := range mod.peers.streams.Clone() {
		if s.id == id {
			return s
		}
	}
	return nil
}

// pickAltStream returns a different stream to the same remote identity as current.
func (mod *Module) pickAltStream(current *Stream) *Stream {
	if current == nil {
		return nil
	}
	for _, s := range mod.peers.streams.Clone() {
		if s == current {
			continue
		}
		if s.RemoteIdentity().IsEqual(current.RemoteIdentity()) {
			return s
		}
	}
	return nil
}

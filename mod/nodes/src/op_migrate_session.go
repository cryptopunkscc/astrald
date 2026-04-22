package nodes

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodesClient "github.com/cryptopunkscc/astrald/mod/nodes/client"
)

type opMigrateSessionArgs struct {
	SessionID astral.Nonce
	StreamID  astral.Nonce
	Buffer    astral.Uint32 `query:"optional"`
	Start     astral.Bool   `query:"optional"`
	Out       string        `query:"optional"`
}

// OpMigrateSession handles session migration.
// With Start=true it acts as the initiator (manual trigger).
// Without Start it acts as the responder (called by the remote initiator over the signalling channel).
func (mod *Module) OpMigrateSession(ctx *astral.Context, q *routing.IncomingQuery, args opMigrateSessionArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if args.Start {
		if args.SessionID == 0 {
			return ch.Send(astral.NewError("missing session_id"))
		}
		if args.StreamID == 0 {
			return ch.Send(astral.NewError("missing stream_id"))
		}

		sess, ok := mod.peers.sessions.Get(args.SessionID)
		if !ok {
			return ch.Send(astral.NewError("session not found"))
		}
		if !sess.IsOpen() {
			return ch.Send(astral.NewError("session not open"))
		}

		targetStream := mod.findStreamByID(args.StreamID)
		if targetStream == nil {
			return ch.Send(astral.NewError("target stream not found"))
		}

		if sess.isOnStream(targetStream) {
			return ch.Send(astral.NewError("session already on target stream"))
		}

		mod.log.Log("migrate session %v to stream %v (manual)", args.SessionID, args.StreamID)

		err := mod.migrateSession(ctx, sess, targetStream)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&astral.Ack{})
	}

	// responder side

	sess, ok := mod.peers.sessions.Get(args.SessionID)
	if !ok {
		return ch.Send(astral.Err(errors.New("session not found")))
	}
	if !sess.IsOpen() {
		return ch.Send(astral.Err(errors.New("session not open")))
	}

	targetStream := mod.findStreamByID(args.StreamID)
	if targetStream == nil {
		return ch.Send(astral.Err(errors.New("target stream not found")))
	}

	migrator := mod.newSessionMigrator(sess, targetStream)
	migrator.prepare()
	err := ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalReady})
	if err != nil {
		return err
	}

	// wait for Migrate frame and drain old input buffer
	if err = migrator.Drain(ctx); err != nil {
		return err
	}

	// send switched
	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalSwitched})
	if err != nil {
		return err
	}

	// wait for resume
	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalResume))
	if err != nil {
		return err
	}

	migrator.Resume(defaultBufferSize)
	mod.log.Logv(1, "session %v migrated to stream %v (responder)", sess.Nonce, targetStream.id)

	return nil
}

// migrateSession runs the initiator side of session migration for a single session.
func (mod *Module) migrateSession(ctx *astral.Context, sess *session, targetStream *Stream) error {
	if !sess.IsOpen() {
		return errors.New("session not open")
	}

	// open signalling channel to remote peer
	ch, err := nodesClient.New(sess.RemoteIdentity, astrald.Default()).MigrateSession(ctx, nodesClient.MigrateSessionArgs{
		SessionID: sess.Nonce,
		StreamID:  targetStream.id,
		Buffer:    astral.Uint32(defaultBufferSize),
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	// wait for ready
	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalReady))
	if err != nil {
		return err
	}

	migrator := mod.newSessionMigrator(sess, targetStream)

	err = migrator.Migrate()
	if err != nil {
		return err
	}

	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalSwitched))
	if err != nil {
		return err
	}

	if err = migrator.Drain(ctx); err != nil {
		return err
	}

	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalResume})
	if err != nil {
		return err
	}

	migrator.Resume(defaultBufferSize)
	mod.log.Logv(1, "session %v migrated to stream %v (initiator)", sess.Nonce, targetStream.id)

	return nil
}

package nodes

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type opMigrateSessionArgs struct {
	SessionID astral.Nonce `query:"required"`
	StreamID  astral.Nonce `query:"required"`
	Start     astral.Bool  `query:"optional"`
	Out       string       `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q *routing.IncomingQuery, args opMigrateSessionArgs) (err error) {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	sess, ok := mod.peers.sessions.Get(args.SessionID)
	if !ok {
		return ch.Send(astral.Err(nodes.ErrSessionNotFound))
	}
	if !sess.IsOpen() {
		return ch.Send(astral.Err(nodes.ErrInvalidSessionState))
	}

	targetStream := mod.findStreamByID(args.StreamID)
	if targetStream == nil {
		return ch.Send(astral.Err(errors.New("target stream not found")))
	}

	if args.Start {
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

	migrator, err := mod.newSessionMigrator(sess)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	var peerBuffer astral.Uint32
	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalReady, &peerBuffer))
	if err != nil {
		return err
	}

	migrator.SetPeerBuffer(int(peerBuffer))

	err = migrator.BeginMigrate(targetStream)
	if err != nil {
		return ch.Send(astral.Err(err))
	}
	defer func() {
		if err != nil {
			migrator.Rollback()
		}
	}()

	err = migrator.SendMigrateFrame()
	if err != nil {
		return err
	}

	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalSwitched, Buffer: astral.Uint32(defaultBufferSize)})
	if err != nil {
		return err
	}

	err = migrator.WaitDrain(ctx)
	if err != nil {
		return err
	}

	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalResume, nil))
	if err != nil {
		return err
	}

	migrator.Resume()
	mod.log.Logv(1, "session %v migrated to stream %v (responder)", sess.Nonce, targetStream.id)

	return nil
}

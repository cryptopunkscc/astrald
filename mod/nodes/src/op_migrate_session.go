package nodes

import (
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

func (mod *Module) OpMigrateSession(ctx *astral.Context, q *routing.IncomingQuery, args opMigrateSessionArgs) error {
	ch := channel.New(q.AcceptRaw(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	stream := mod.findStreamBySessionNonce(args.SessionID)
	if stream == nil {
		return ch.Send(astral.Err(nodes.ErrSessionNotFound))
	}
	session, ok := stream.Mux.sessions.Get(args.SessionID)
	if !ok {
		return ch.Send(astral.Err(nodes.ErrSessionNotFound))
	}
	if !session.IsOpen() {
		return ch.Send(astral.Err(nodes.ErrInvalidSessionState))
	}

	targetStream := mod.findStreamByID(args.StreamID)
	if targetStream == nil {
		return ch.Send(astral.Err(nodes.ErrStreamNotFound))
	}

	if args.Start {
		if session.isOnStream(targetStream) {
			return ch.Send(astral.NewError("session already on target stream"))
		}

		mod.log.Log("migrate session %v to stream %v (manual)", args.SessionID, args.StreamID)
		migrateCtx, cancel := ctx.WithTimeout(migrateSessionTimeout)
		defer cancel()
		err := mod.migrateSession(migrateCtx, session, targetStream)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		return ch.Send(&astral.Ack{})
	}

	migrator, err := mod.newSessionMigrator(session)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	var peerBuffer astral.Uint32
	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalReady, &peerBuffer))
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	migrator.SetPeerBuffer(int(peerBuffer))
	err = migrator.Begin(targetStream)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// note: submethod made for sake of defer (im thinking about other solution)
	migrateCtx, cancel := ctx.WithTimeout(migrateSessionTimeout)
	defer cancel()
	err = mod.respondMigration(migrateCtx, ch, session, migrator)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	mod.log.Logv(1, "session %v migrated to stream %v (responder)", session.Nonce, targetStream.id)
	return nil
}

func (mod *Module) respondMigration(ctx *astral.Context, ch *channel.Channel, session *session, migrator *SessionMigrator) (err error) {
	defer func() {
		if err != nil {
			session.Close()
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

	err = migrator.WaitClosed(ctx)
	if err != nil {
		return err
	}

	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalResume, nil))
	if err != nil {
		return err
	}

	err = migrator.Complete()
	if err != nil {
		return err
	}

	return ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalDone})
}

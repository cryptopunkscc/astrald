package nodes

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodesClient "github.com/cryptopunkscc/astrald/mod/nodes/client"
)

const migrationTotalTimeout = 10 * time.Second

type opMigrateSessionArgs struct {
	SessionID astral.Nonce
	StreamID  astral.Nonce
	Start     astral.Bool `query:"optional"`
	Out       string      `query:"optional"`
}

func (mod *Module) OpMigrateSession(ctx *astral.Context, q *ops.Query, args opMigrateSessionArgs) error {
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	if args.SessionID == 0 {
		return ch.Send(astral.NewError("missing sessionId"))
	}
	if args.StreamID == 0 {
		return ch.Send(astral.NewError("missing stream ids"))
	}

	ctx, cancel := ctx.WithTimeout(migrationTotalTimeout)
	defer cancel()

	session, ok := mod.peers.sessions.Get(args.SessionID)
	if !ok {
		return ch.Send(astral.NewError("session not found"))
	}

	migrator, err := mod.createSessionMigrator(session, args.StreamID)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	if args.Start {
		client := nodesClient.New(session.RemoteIdentity, astrald.Default())
		if err := client.MigrateSession(ctx, args.SessionID, args.StreamID, migrator); err != nil {
			return ch.Send(astral.Err(err))
		}

		mod.log.Log("migration: session migrated %v %v", args.SessionID, args.StreamID)
		return ch.Send(&astral.Ack{})
	}

	if err := ch.Switch(
		nodes.ExpectMigrateSignal(args.SessionID, nodes.MigrateSignalTypeBegin),
		channel.PassErrors,
		channel.WithContext(ctx),
	); err != nil {
		return ch.Send(astral.Err(err))
	}

	if err := migrator.Migrate(); err != nil {
		return ch.Send(astral.Err(err))
	}

	defer session.CancelMigration()

	if err := ch.Send(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeReady, Nonce: args.SessionID}); err != nil {
		return err
	}

	if err := session.WaitOpen(ctx); err != nil {
		return ch.Send(astral.Err(err))
	}

	if err := ch.Send(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeCompleted, Nonce: args.SessionID}); err != nil {
		return err
	}

	mod.log.Log("migration: session migrated %v %v", args.SessionID, args.StreamID)
	return ch.Send(&astral.Ack{})
}

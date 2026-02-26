package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (client *Client) MigrateSession(ctx *astral.Context, sessionID, streamID astral.Nonce, m nodes.SessionMigrator) error {
	defer m.CancelMigration()

	ch, err := client.queryCh(ctx, nodes.MethodMigrateSession, query.Args{
		"session_id": sessionID,
		"stream_id":  streamID,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Send(&nodes.SessionMigrateSignal{Signal: nodes.MigrateSignalTypeBegin, Nonce: sessionID}); err != nil {
		return err
	}
	if err := ch.Switch(
		nodes.ExpectMigrateSignal(sessionID, nodes.MigrateSignalTypeReady),
		channel.PassErrors,
		channel.WithContext(ctx),
	); err != nil {
		return err
	}
	if err := m.Migrate(); err != nil {
		return err
	}
	if err := m.WriteMigrateFrame(); err != nil {
		return err
	}

	return m.WaitOpen(ctx)
}

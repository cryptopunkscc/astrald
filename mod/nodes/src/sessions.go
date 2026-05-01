package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodesClient "github.com/cryptopunkscc/astrald/mod/nodes/client"
)

// migrateSession migrates a single session to targetStream (initiator side).
func (mod *Module) migrateSession(ctx *astral.Context, session *session, targetStream *Link) (err error) {
	ch, err := nodesClient.New(session.RemoteIdentity, astrald.Default()).MigrateSession(ctx, nodesClient.MigrateSessionArgs{
		SessionID: session.Nonce,
		StreamID:  targetStream.id,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	migrator, err := mod.newSessionMigrator(session)
	if err != nil {
		return err
	}

	err = migrator.Begin(targetStream)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			session.Close()
		}
	}()

	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalReady, Buffer: astral.Uint32(defaultBufferSize)})
	if err != nil {
		return err
	}

	var peerBuffer astral.Uint32
	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalSwitched, &peerBuffer))
	if err != nil {
		return err
	}

	migrator.SetPeerBuffer(int(peerBuffer))
	err = migrator.SendMigrateFrame()
	if err != nil {
		return err
	}

	err = migrator.WaitClosed(ctx)
	if err != nil {
		return err
	}

	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalResume})
	if err != nil {
		return err
	}

	err = ch.Switch(nodes.ExpectMigrateSignal(nodes.MigrateSignalDone, nil))
	if err != nil {
		return err
	}

	err = migrator.Complete()
	if err != nil {
		return err
	}

	mod.log.Logv(1, "session %v migrated to stream %v (initiator)", session.Nonce, targetStream.id)
	return nil
}

package nodes

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	nodesClient "github.com/cryptopunkscc/astrald/mod/nodes/client"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

func (mod *Peers) newSession(nonce astral.Nonce, remoteID *astral.Identity, queryStr string) (*session, bool) {
	sess, ok := mod.sessions.Set(nonce, newSession(nonce))
	if !ok {
		return sess, false
	}
	sess.RemoteIdentity = remoteID
	sess.Query = queryStr
	return sess, true
}

func (mod *Peers) newMuxInputBuffer(s *Stream, nonce astral.Nonce) *InputBuffer {
	onRead := func(n int) {
		s.Write(&frames.Read{Nonce: nonce, Len: uint32(n)})
	}

	return NewInputBuffer(defaultBufferSize, onRead)
}

func (mod *Peers) newMuxOutputBuffer(s *Stream, nonce astral.Nonce, sess *session) *OutputBuffer {
	onWrite := func(p []byte) error {
		remaining := p
		for len(remaining) > 0 {
			chunkSize := maxPayloadSize
			if len(remaining) < chunkSize {
				chunkSize = len(remaining)
			}

			chunk := remaining[:chunkSize]

			sess.bytes.Add(uint64(len(chunk)))

			if err := s.Write(&frames.Data{
				Nonce:   nonce,
				Payload: chunk,
			}); err != nil {
				return err
			}

			remaining = remaining[chunkSize:]
		}

		return nil
	}

	return NewOutputBuffer(onWrite)
}

// migrateSession migrates single session
func (mod *Module) migrateSession(ctx *astral.Context, session *session, targetStream *Stream) error {
	if !session.IsOpen() {
		return errors.New("session not open")
	}

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

	err = migrator.BeginMigrate(targetStream)
	if err != nil {
		return err
	}

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

	err = migrator.WaitDrain(ctx)
	if err != nil {
		return err
	}

	err = ch.Send(&nodes.MigrateSignal{Signal: nodes.MigrateSignalResume})
	if err != nil {
		return err
	}

	migrator.Resume()
	mod.log.Logv(1, "session %v migrated to stream %v (initiator)", session.Nonce, targetStream.id)
	return nil
}

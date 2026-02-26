package nodes

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type sessionMigrator struct {
	sess         *session
	targetStream *Stream
}

var _ nodes.SessionMigrator = &sessionMigrator{}

func (m *sessionMigrator) Migrate() error           { return m.sess.Migrate(m.targetStream) }
func (m *sessionMigrator) WriteMigrateFrame() error { return m.sess.writeMigrateFrame() }
func (m *sessionMigrator) CompleteMigration() error { return m.sess.CompleteMigration() }
func (m *sessionMigrator) CancelMigration()         { m.sess.CancelMigration() }

func (mod *Module) createSessionMigrator(sess *session, streamID astral.Nonce) (nodes.SessionMigrator, error) {
	targetStream := mod.findStreamByID(streamID)
	if targetStream == nil {
		return nil, errors.New("target stream not found")
	}
	return &sessionMigrator{sess: sess, targetStream: targetStream}, nil
}

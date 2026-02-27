package nodes

import (
	"context"
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type sessionMigrator struct {
	session      *session
	targetStream *Stream
}

var _ nodes.SessionMigrator = &sessionMigrator{}

func (m *sessionMigrator) Migrate() error                     { return m.session.Migrate(m.targetStream) }
func (m *sessionMigrator) WriteMigrateFrame() error           { return m.session.writeMigrateFrame() }
func (m *sessionMigrator) CancelMigration()                   { m.session.CancelMigration() }
func (m *sessionMigrator) WaitOpen(ctx context.Context) error { return m.session.WaitOpen(ctx) }

func (mod *Module) createSessionMigrator(sess *session, streamID astral.Nonce) (nodes.SessionMigrator, error) {
	targetStream := mod.findStreamByID(streamID)
	if targetStream == nil {
		return nil, errors.New("target stream not found")
	}
	return &sessionMigrator{session: sess, targetStream: targetStream}, nil
}

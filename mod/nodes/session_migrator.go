package nodes

import "context"

// SessionMigrator abstracts session migration operations for use in the client protocol.
type SessionMigrator interface {
	Migrate() error
	WriteMigrateFrame() error
	CancelMigration()
	WaitOpen(ctx context.Context) error
}

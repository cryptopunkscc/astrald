package nodes

// SessionMigrator abstracts session migration operations for use in the client protocol.
type SessionMigrator interface {
	Migrate() error
	WriteMigrateFrame() error
	CompleteMigration() error
	CancelMigration()
}

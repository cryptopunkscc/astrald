package storage

func (mod *Module) setupDatabase() (err error) {
	// Migrate the schema
	if err := mod.db.AutoMigrate(&dbAccess{}); err != nil {
		return err
	}

	return nil
}

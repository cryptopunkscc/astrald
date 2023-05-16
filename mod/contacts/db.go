package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
	"time"
)

const DatabaseName = "contacts.db"

func (m *Module) setupDatabase() (err error) {
	var path = filepath.Join(m.rootDir, DatabaseName)

	m.db, err = gorm.Open(
		sqlite.Open(path),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return err
	}

	// Migrate the schema
	if err := m.db.AutoMigrate(&dbNode{}); err != nil {
		return err
	}

	return nil
}

type dbNode struct {
	CreatedAt time.Time
	Identity  string `gorm:"primaryKey"`
	Alias     string `gorm:"index"`
}

func (dbNode) TableName() string {
	return "nodes"
}

type Node struct {
	Identity id.Identity
	Alias    string
}

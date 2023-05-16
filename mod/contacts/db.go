package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
	"time"
)

func (m *Module) setupDatabase() (err error) {
	var path = DatabaseName

	if coreNode, ok := m.node.(*node.CoreNode); ok {
		path = filepath.Join(coreNode.RootDir(), DatabaseName)
	}

	m.db, err = gorm.Open(
		sqlite.Open(path),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		return err
	}

	// Migrate the schema
	if err := m.db.AutoMigrate(&DbNode{}); err != nil {
		return err
	}

	return nil
}

type DbNode struct {
	CreatedAt time.Time
	Identity  string `gorm:"primaryKey"`
	Alias     string `gorm:"index"`
}

type Node struct {
	Identity id.Identity
	Alias    string
}

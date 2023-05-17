package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

func (m *Module) setupDatabase() (err error) {
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

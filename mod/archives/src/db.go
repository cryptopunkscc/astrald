package archives

import (
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbArchive struct {
	ID        uint       `gorm:"primarykey"`
	ObjectID  *object.ID `gorm:"uniqueIndex"`
	Entries   []dbEntry  `gorm:"OnDelete:CASCADE;foreignKey:ParentID"`
	Format    string     `gorm:"index"`
	Comment   string
	CreatedAt time.Time
}

func (dbArchive) TableName() string { return archives.DBPrefix + "archives" }

type dbEntry struct {
	Parent   *dbArchive
	ParentID uint       `gorm:"primaryKey"`
	Path     string     `gorm:"primaryKey"`
	ObjectID *object.ID `gorm:"index"`
	Comment  string
	Modified time.Time
}

func (dbEntry) TableName() string { return archives.DBPrefix + "entries" }

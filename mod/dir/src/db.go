package dir

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

type dbAlias struct {
	Identity id.Identity `gorm:"primaryKey"`
	Alias    string      `gorm:"index;unique;not null"`
}

func (dbAlias) TableName() string {
	return dir.DBPrefix + "aliases"
}

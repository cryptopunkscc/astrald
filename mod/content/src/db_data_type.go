package content

import (
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbDataType struct {
	DataID       object.ID `gorm:"primaryKey"`
	Type         string    `gorm:"index"`
	Method       string    `gorm:"index"`
	IdentifiedAt time.Time `gorm:"index"`
}

func (dbDataType) TableName() string {
	return content.DBPrefix + "data_types"
}

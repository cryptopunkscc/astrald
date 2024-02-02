package content

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"time"
)

type dbDataType struct {
	DataID       data.ID   `gorm:"primaryKey,index"`
	Type         string    `gorm:"index"`
	Method       string    `gorm:"index"`
	IdentifiedAt time.Time `gorm:"index"`
}

func (dbDataType) TableName() string {
	return content.DBPrefix + "data_types"
}

package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type dbRelayCert struct {
	DataID    data.ID     `gorm:"primaryKey"`
	TargetID  id.Identity `gorm:"index"`
	RelayID   id.Identity `gorm:"index"`
	Direction string      `gorm:"index"`
	ExpiresAt time.Time   `gorm:"index"`
}

func (dbRelayCert) TableName() string {
	return "relay_certs"
}

func (mod *Module) dbFindByDataID(dataID data.ID) (*dbRelayCert, error) {
	var row dbRelayCert

	var tx = mod.db.Where("data_id = ?", dataID.String()).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

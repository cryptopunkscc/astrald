package keys

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

type dbPrivateKey struct {
	DataID    data.ID     `gorm:"uniqueIndex"`
	Type      string      `gorm:"index"`
	PublicKey id.Identity `gorm:"index"`
}

func (dbPrivateKey) TableName() string {
	return "private_keys"
}

func (mod *Module) dbFindByDataID(id data.ID) (*dbPrivateKey, error) {
	var row dbPrivateKey

	var tx = mod.db.Where("data_id = ?", id.String()).First(&row)

	return &row, tx.Error
}

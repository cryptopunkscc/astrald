package keys

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

type dbPrivateKey struct {
	DataID    data.ID     `gorm:"uniqueIndex"`
	Type      string      `gorm:"index"`
	PublicKey id.Identity `gorm:"index"`
}

func (dbPrivateKey) TableName() string {
	return keys.DBPrefix + "private_keys"
}

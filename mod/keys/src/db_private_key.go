package keys

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/object"
)

type dbPrivateKey struct {
	DataID    object.ID   `gorm:"uniqueIndex"`
	Type      string      `gorm:"index"`
	PublicKey id.Identity `gorm:"index"`
}

func (dbPrivateKey) TableName() string {
	return keys.DBPrefix + "private_keys"
}

package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/keys"
)

type dbPrivateKey struct {
	DataID    *astral.ObjectID `gorm:"uniqueIndex"`
	Type      string           `gorm:"index"`
	PublicKey *astral.Identity `gorm:"index"`
}

func (dbPrivateKey) TableName() string {
	return keys.DBPrefix + "private_keys"
}

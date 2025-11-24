package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type dbNodeContractRevocation struct {
	ObjectID   *astral.ObjectID `gorm:"primaryKey"`
	UserID     *astral.Identity `gorm:"index"`
	ContractID *astral.ObjectID `gorm:"index"`
	StartsAt   time.Time        `gorm:"index"`
	ExpiresAt  time.Time        `gorm:"index"`
}

func (dbNodeContractRevocation) TableName() string {
	return user.DBPrefix + "node_contract_revocations"
}

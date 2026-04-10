package auth

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"gorm.io/gorm"
)

type DB struct{ *gorm.DB }

type dbContract struct {
	ObjectID  *astral.ObjectID `gorm:"primaryKey"`
	IssuerID  *astral.Identity `gorm:"index"`
	SubjectID *astral.Identity `gorm:"index"`
	StartsAt  time.Time
	ExpiresAt time.Time
}

func (dbContract) TableName() string { return auth.DBPrefix + "contracts" }

type dbContractRevocation struct {
	ObjectID   *astral.ObjectID `gorm:"primaryKey"`
	ContractID *astral.ObjectID `gorm:"index"`
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

func (dbContractRevocation) TableName() string { return auth.DBPrefix + "contract_revocations" }

func (db *DB) contractRevocationExists(objectID *astral.ObjectID) bool {
	var count int64
	db.Model(&dbContractRevocation{}).Where("object_id = ?", objectID).Count(&count)
	return count > 0
}

func (db *DB) findContractRevocationsByContract(contractID *astral.ObjectID) ([]*dbContractRevocation, error) {
	var rows []*dbContractRevocation
	return rows, db.Where("contract_id = ?", contractID).Find(&rows).Error
}

func (db *DB) findActiveContractsBySubject(subjectID *astral.Identity) ([]*dbContract, error) {
	var rows []*dbContract
	now := time.Now()
	return rows, db.
		Where("subject_id = ?", subjectID).
		Where("starts_at <= ?", now).
		Where("expires_at = ? OR expires_at > ?", time.Time{}, now).
		Find(&rows).Error
}

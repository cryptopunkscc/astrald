package auth

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type dbBan struct {
	SubjectID *astral.Identity `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (dbBan) TableName() string { return auth.DBPrefix + "bans" }

func (db *DB) findActiveContractsBySubject(subjectID *astral.Identity) ([]*dbContract, error) {
	var rows []*dbContract
	now := time.Now()
	return rows, db.
		Where("subject_id = ?", subjectID).
		Where("starts_at <= ?", now).
		Where("expires_at = ? OR expires_at > ?", time.Time{}, now).
		Find(&rows).Error
}

func (db *DB) findActiveContractsByIssuer(issuerID *astral.Identity) ([]*dbContract, error) {
	var rows []*dbContract
	now := time.Now()
	return rows, db.
		Where("issuer_id = ?", issuerID).
		Where("starts_at <= ?", now).
		Where("expires_at = ? OR expires_at > ?", time.Time{}, now).
		Find(&rows).Error
}

func (db *DB) storeContract(objectID *astral.ObjectID, issuerID, subjectID *astral.Identity, expiresAt time.Time) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbContract{
			ObjectID:  objectID,
			IssuerID:  issuerID,
			SubjectID: subjectID,
			ExpiresAt: expiresAt,
		}).Error
}

func (db *DB) addBan(subjectID *astral.Identity) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbBan{SubjectID: subjectID, CreatedAt: time.Now()}).Error
}

func (db *DB) removeBan(subjectID *astral.Identity) error {
	return db.Where("subject_id = ?", subjectID).Delete(&dbBan{}).Error
}

func (db *DB) isBanned(subjectID *astral.Identity) bool {
	var count int64
	db.Model(&dbBan{}).Where("subject_id = ?", subjectID).Count(&count)
	return count > 0
}

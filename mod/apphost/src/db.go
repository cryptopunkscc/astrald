package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type DB struct {
	*gorm.DB
}

func (db *DB) CreateAccessToken(identity *astral.Identity, d astral.Duration) (token *dbAccessToken, err error) {
	var expiresAt = (astral.Time)(time.Now().Add(time.Duration(d)))

	token = &dbAccessToken{
		Identity:  identity,
		Token:     randomString(32),
		ExpiresAt: time.Time(expiresAt),
	}

	err = db.Create(token).Error

	return
}

func (db *DB) ListAccessTokens() (list []dbAccessToken, _ error) {
	return list, db.Find(&list).Error
}

func (db *DB) FindAccessToken(token string) (at *dbAccessToken, err error) {
	err = db.
		Where("token = ?", token).
		First(&at).Error
	return
}

func (db *DB) FindAppContract(id *object.ID) (ac *dbAppContract, err error) {
	err = db.
		Where("object_id = ?", id).
		First(&ac).Error
	return
}

func (db *DB) SaveAppContract(ac *dbAppContract) (err error) {
	return db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(ac).Error
}

func (db *DB) FindActiveAppContractsByHost(hostID *astral.Identity) (rows []*dbAppContract, err error) {
	now := time.Now()

	err = db.
		Where("starts_at <= ? AND expires_at >= ?", now, now).
		Where("host_id = ?", hostID).
		Find(&rows).Error

	return
}

func (db *DB) FindActiveAppContractsByApp(appID *astral.Identity) (rows []*dbAppContract, err error) {
	now := time.Now()

	err = db.
		Where("starts_at <= ? AND expires_at >= ?", now, now).
		Where("app_id = ?", appID).
		Find(&rows).Error

	return
}

func (db *DB) FindActiveAppContractsByAppAndHost(appId, hostID *astral.Identity) (rows []*dbAppContract, err error) {
	now := time.Now()

	err = db.
		Model(&dbAppContract{}).
		Where("starts_at <= ? AND expires_at >= ?", now, now).
		Where("app_id = ? AND host_id = ?", appId, hostID).
		Order("expires_at DESC").
		Find(&rows).Error

	return
}

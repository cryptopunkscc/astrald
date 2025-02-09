package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
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
	if err != nil {
		at = nil
	}
	return
}

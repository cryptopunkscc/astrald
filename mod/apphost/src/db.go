package apphost

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (db *DB) CreateLocalApp(appID, hostID *astral.Identity) error {
	row := &dbLocalApp{AppID: appID, HostID: hostID, InstalledAt: time.Now()}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error
}

func (db *DB) ListLocalApps() (list []*dbLocalApp, err error) {
	return list, db.Find(&list).Error
}

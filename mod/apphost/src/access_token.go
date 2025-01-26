package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"math/rand"
)

type dbAccessToken struct {
	Identity *astral.Identity `gorm:"index"`
	Token    string           `gorm:"uniqueIndex"`
}

func (dbAccessToken) TableName() string {
	return apphost.DBPrefix + "access_tokens"
}

func (mod *Module) CreateAccessToken(identity *astral.Identity) (string, error) {
	var token = randomString(32)

	var tx = mod.db.Create(&dbAccessToken{
		Identity: identity,
		Token:    token,
	})

	return token, tx.Error
}

func (mod *Module) FindAccessToken(identity *astral.Identity) (token string, err error) {
	err = mod.db.
		Model(&dbAccessToken{}).
		Where("identity = ?", identity).
		Select("token").
		First(&token).Error
	return
}

func (mod *Module) FindOrCreateAccessToken(identity *astral.Identity) (token string, err error) {
	token, err = mod.FindAccessToken(identity)
	if err == nil {
		return
	}
	return mod.CreateAccessToken(identity)
}

func (mod *Module) identityByToken(token string) (identity *astral.Identity) {
	if len(token) == 0 {
		return
	}

	mod.db.
		Model(&dbAccessToken{}).
		Where("token = ?", token).
		Select("identity").
		First(&identity)

	return
}

func randomString(length int) (s string) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	var name = make([]byte, length)
	for i := 0; i < len(name); i++ {
		name[i] = charset[rand.Intn(len(charset))]
	}
	return string(name[:])
}

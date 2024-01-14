package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"math/rand"
)

type dbAccessToken struct {
	Identity string `gorm:"index"`
	Token    string `gorm:"uniqueIndex"`
}

func (dbAccessToken) TableName() string {
	return "access_tokens"
}

func (mod *Module) CreateAccessToken(identity id.Identity) (string, error) {
	var token = randomString(32)

	var tx = mod.db.Create(&dbAccessToken{
		Identity: identity.PublicKeyHex(),
		Token:    token,
	})

	return token, tx.Error
}

func (mod *Module) authToken(token string) (identity id.Identity) {
	var row dbAccessToken

	var tx = mod.db.Where("token = ?", token).First(&row)
	if tx.Error != nil {
		return
	}

	identity, _ = id.ParsePublicKeyHex(row.Identity)

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

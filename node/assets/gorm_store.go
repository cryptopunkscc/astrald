package assets

import (
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"gorm.io/gorm"
	"time"
)

type GormKeyStore struct {
	db *gorm.DB
}

func NewGormKeyStore(db *gorm.DB) (*GormKeyStore, error) {
	var err error
	var store = &GormKeyStore{
		db: db,
	}
	if err != nil {
		return nil, err
	}

	err = store.migrateDB()
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (store *GormKeyStore) Save(identity id.Identity) error {
	if identity.PrivateKey() == nil {
		return errors.New("private key missing")
	}

	return store.db.Create(&gormIdentity{
		PublicKey:  identity.PublicKeyHex(),
		PrivateKey: hex.EncodeToString(identity.PrivateKey().Serialize()),
	}).Error
}

func (store *GormKeyStore) Find(identity id.Identity) (id.Identity, error) {
	var record gormIdentity

	if tx := store.db.First(&record, "public_key = ?", identity.PublicKeyHex()); tx.Error != nil {
		return id.Identity{}, tx.Error
	}

	return record.Identity(), nil
}

func (store *GormKeyStore) migrateDB() error {
	return store.db.AutoMigrate(&gormIdentity{})
}

func (i gormIdentity) Identity() id.Identity {
	if i.PrivateKey != "" {
		if pk, err := hex.DecodeString(i.PrivateKey); err == nil {
			if identity, err := id.ParsePrivateKey(pk); err == nil {
				return identity
			}
		}
	}

	if i.PublicKey != "" {
		if identity, err := id.ParsePublicKeyHex(i.PublicKey); err == nil {
			return identity
		}
	}

	return id.Identity{}
}

type gormIdentity struct {
	PublicKey  string `gorm:"index"`
	PrivateKey string
	CreatedAt  time.Time
}

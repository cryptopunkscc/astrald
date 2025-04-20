package user

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type DB struct {
	mod *Module
	*gorm.DB
}

func (db *DB) UniqueActiveUsersOnNode(nodeID *astral.Identity) (users []*astral.Identity, err error) {
	err = db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("node_id = ?", nodeID).
		Distinct("user_id").
		Find(&users).
		Error

	return
}

func (db *DB) UniqueActiveNodesOfUser(userID *astral.Identity) (nodes []*astral.Identity, err error) {
	err = db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("user_id = ?", userID).
		Distinct("node_id").
		Find(&nodes).
		Error

	return
}

func (db *DB) ActiveContractsOf(userID *astral.Identity) (contracts []*dbNodeContract, err error) {
	err = db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("user_id = ?", userID).
		Find(&contracts).
		Error

	return
}

func (db *DB) ContractExists(contractID object.ID) (b bool) {
	db.
		Model(&dbNodeContract{}).
		Where("object_id = ?", contractID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (db *DB) Contacts() (contacts []*astral.Identity) {
	db.
		Model(&dbContact{}).
		Select("user_id").
		Find(&contacts)
	return
}

func (db *DB) IsContact(userID *astral.Identity) (b bool) {
	if userID.IsZero() {
		return
	}
	db.
		Model(&dbContact{}).
		Where("user_id = ?", userID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (db *DB) AddContact(userID *astral.Identity) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContact{
		UserID: userID,
	}).Error
}

func (db *DB) RemoveContact(userID *astral.Identity) error {
	return db.Delete(&dbContact{UserID: userID}).Error
}

func (db *DB) AssetHeight() int {
	var height uint64

	err := db.Model(&dbAsset{}).
		Select("MAX(height) AS height").
		Scan(&height).
		Error

	if err != nil {
		return -1
	}

	return int(height)
}

func (db *DB) AssetsContain(objectID *object.ID) (exists bool) {
	db.Model(&dbAsset{}).
		Select("1").
		Where("object_id = ? and removed = false", objectID).
		Limit(1).
		Scan(&exists)
	return
}

func (db *DB) AddAsset(objectID *object.ID, force bool) (nonce astral.Nonce, err error) {
	if !force && db.AssetsContain(objectID) {
		return nonce, errors.New("asset already exists")
	}

	nonce = astral.NewNonce()

	return nonce, db.Save(&dbAsset{
		Nonce:    nonce,
		ObjectID: objectID,
		Height:   uint64(db.AssetHeight() + 1),
	}).Error
}

func (db *DB) RemoveAsset(objectID *object.ID) (err error) {
	var rows []dbAsset

	err = db.Where("object_id = ? and removed = false", objectID).Find(&rows).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		row.Removed = true
		row.Height = uint64(db.AssetHeight() + 1)
		err = db.Save(row).Error
		if err != nil {
			return
		}
	}

	return
}

func (db *DB) Assets() (assets []*object.ID, err error) {
	err = db.
		Model(&dbAsset{}).
		Where("removed = false").
		Distinct("object_id").
		Find(&assets).
		Error

	return
}

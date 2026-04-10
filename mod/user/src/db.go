package user

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
)

type DB struct {
	mod *Module
	*gorm.DB
}

func (db *DB) assetExists(objectID *astral.ObjectID) (exists bool) {
	db.Model(&dbAsset{}).
		Select("1").
		Where("object_id = ? and removed = false", objectID).
		Limit(1).
		Scan(&exists)
	return
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

func (db *DB) NonceExists(nonce astral.Nonce) (exists bool) {
	db.Model(&dbAsset{}).
		Select("1").
		Where("nonce", nonce).
		Limit(1).
		Scan(&exists)
	return
}

func (db *DB) AddAsset(objectID *astral.ObjectID, force bool) (nonce astral.Nonce, err error) {
	if !force && db.assetExists(objectID) {
		return nonce, errors.New("asset already exists")
	}

	nonce = astral.NewNonce()

	return nonce, db.Save(&dbAsset{
		Nonce:    nonce,
		ObjectID: objectID,
		Height:   uint64(db.AssetHeight() + 1),
	}).Error
}

func (db *DB) AddAssetWithNonce(objectID *astral.ObjectID, nonce astral.Nonce) error {
	if db.NonceExists(nonce) {
		return nil
	}

	return db.Create(&dbAsset{
		Nonce:    nonce,
		ObjectID: objectID,
		Height:   uint64(db.AssetHeight() + 1),
	}).Error
}

func (db *DB) RemoveAsset(objectID *astral.ObjectID) (err error) {
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

func (db *DB) RemoveAssetByNonce(nonce astral.Nonce, objectID *astral.ObjectID) (err error) {
	var row dbAsset

	err = db.Where("nonce = ?", nonce).First(&row).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&dbAsset{
			Nonce:    nonce,
			Removed:  true,
			ObjectID: objectID,
			Height:   uint64(db.AssetHeight() + 1),
		}).Error

	case err != nil:
		return
	}

	if row.Removed {
		return nil
	}

	row.Removed = true
	row.Height = uint64(db.AssetHeight() + 1)

	return db.Save(row).Error
}

func (db *DB) Assets() (assets []*astral.ObjectID, err error) {
	err = db.
		Model(&dbAsset{}).
		Where("removed = false").
		Distinct("object_id").
		Find(&assets).
		Error

	return
}

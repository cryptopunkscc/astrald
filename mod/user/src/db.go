package user

import (
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
	"gorm.io/gorm"
)

type DB struct {
	mod *Module
	*gorm.DB
}

func (db *DB) findSignedNodeContract(contractID *astral.ObjectID) (row *dbSignedNodeContract, err error) {
	err = db.
		Model(&dbSignedNodeContract{}).
		Where("object_id = ?", contractID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrContractNotExists
		}
		return nil, err
	}

	return row, nil
}

func (db *DB) deleteSignedNodeContract(contractID *astral.ObjectID) error {
	return db.Delete(&dbSignedNodeContract{}, "object_id = ?", contractID).Error
}

func (db *DB) signedNodeContractExists(contractID *astral.ObjectID) (b bool) {
	db.
		Model(&dbSignedNodeContract{}).
		Where("object_id = ?", contractID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (db *DB) findNodeContractRevocation(revocationID *astral.ObjectID) (row *dbNodeContractRevocation, err error) {
	err = db.
		Model(&dbNodeContractRevocation{}).
		Where("object_id = ?", revocationID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, user.ErrContractRevocationNotExists
		}
		return nil, err
	}

	return row, nil
}

func (db *DB) nodeContractRevocationExists(revocationID *astral.ObjectID) (b bool) {
	db.
		Model(&dbNodeContractRevocation{}).
		Where("object_id = ?", revocationID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (db *DB) assetExists(objectID *astral.ObjectID) (exists bool) {
	db.Model(&dbAsset{}).
		Select("1").
		Where("object_id = ? and removed = false", objectID).
		Limit(1).
		Scan(&exists)
	return
}

func (db *DB) UniqueActiveUsersOnNode(nodeID *astral.Identity) (users []*astral.Identity, err error) {
	now := time.Now().UTC()

	err = db.
		Model(&dbSignedNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("node_id = ?", nodeID).
		Where(`
			NOT EXISTS (
				SELECT 1 FROM users__node_contract_revocations r
				WHERE r.contract_id = users__signed_node_contracts.object_id
				  AND r.expires_at > ?
			)
		`, now, now).
		Distinct("user_id").
		Find(&users).
		Error

	return
}

func (db *DB) UniqueActiveNodesOfUser(userID *astral.Identity) (nodes []*astral.Identity, err error) {
	now := time.Now().UTC()

	err = db.
		Model(&dbSignedNodeContract{}).
		Where("starts_at < ?", now).
		Where("expires_at > ?", now).
		Where("user_id = ?", userID).
		Where(`
			NOT EXISTS (
				SELECT 1 FROM users__node_contract_revocations r
				WHERE r.contract_id = users__signed_node_contracts.object_id
				  AND r.expires_at > ?
			)
		`, now, now).
		Distinct("node_id").
		Find(&nodes).
		Error

	return
}

func (db *DB) ActiveContractsOf(userID *astral.Identity) (contracts []*dbSignedNodeContract, err error) {
	now := time.Now().UTC()

	err = db.
		Model(&dbSignedNodeContract{}).
		Where("starts_at < ?", now).
		Where("expires_at > ?", now).
		Where("user_id = ?", userID).
		Where(`
			NOT EXISTS (
				SELECT 1 FROM users__node_contract_revocations r
				WHERE r.contract_id = users__signed_node_contracts.object_id
				  AND r.expires_at > ?
			)
		`, now, now).
		Find(&contracts).
		Error

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

func (db *DB) ContractRevocationExists(revocationID *astral.ObjectID) (b bool) {
	db.
		Model(&dbNodeContractRevocation{}).
		Where("object_id = ?", revocationID).
		Select("count(*) > 0").
		First(&b)

	return
}

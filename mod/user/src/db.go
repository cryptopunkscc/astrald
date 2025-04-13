package user

import (
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

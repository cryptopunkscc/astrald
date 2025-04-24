package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	*gorm.DB
}

func (db *DB) AddEndpoint(nodeID *astral.Identity, network, address string) error {
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbEndpoint{
			Identity: nodeID,
			Network:  network,
			Address:  address,
		}).Error
}

func (db *DB) RemoveEndpoint(nodeID *astral.Identity, network, address string) error {
	return db.Delete(&dbEndpoint{
		Identity: nodeID,
		Network:  network,
		Address:  address,
	}).Error
}

func (db *DB) FindEndpoints(nodeID *astral.Identity) (rows []*dbEndpoint, err error) {
	err = db.Find(&rows, "identity = ?", nodeID).Error

	return
}

func (db *DB) HasEndpoints(nodeID *astral.Identity) (has bool) {
	db.
		Model(&dbEndpoint{}).
		Where("identity = ?", nodeID).
		Select("count(*) > 0").
		First(&has)
	return
}

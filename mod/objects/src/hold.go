package objects

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm/clause"
)

func (mod *Module) Hold(identity id.Identity, objectIDs ...object.ID) error {
	var rows []dbHolding

	for _, objectID := range objectIDs {
		rows = append(rows, dbHolding{
			HolderID: identity,
			ObjectID: objectID,
		})
	}

	err := mod.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	if err == nil {
		mod.events.Emit(objects.EventHeld{
			HolderID:  identity,
			ObjectIDs: objectIDs,
		})
	}
	return err
}

func (mod *Module) Release(identity id.Identity, objectIDs ...object.ID) error {
	err := mod.db.
		Where("holder_id = ? and object_id IN (?)", identity, objectIDs).
		Delete(&dbHolding{}).
		Error
	if err == nil {
		mod.events.Emit(objects.EventReleased{
			HolderID:  identity,
			ObjectIDs: objectIDs,
		})
	}
	return err
}

func (mod *Module) Holders(objectID object.ID) (holders []id.Identity) {
	err := mod.db.
		Model(&dbHolding{}).
		Where("object_id = ?", objectID).
		Select("holder_id").
		Find(&holders).
		Error

	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

func (mod *Module) Holdings(identity id.Identity) (holdings []object.ID) {
	err := mod.db.
		Model(&dbHolding{}).
		Where("holder_id = ?", identity).
		Select("object_id").
		Find(&holdings).
		Error

	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return
}

func (mod *Module) isHolding(holderID id.Identity, objectID object.ID) bool {
	var c int64
	err := mod.db.Model(&dbHolding{}).Where("holder_id = ? and object_id = ?", holderID, objectID).Count(&c)

	if err != nil {
		mod.log.Error("db error: %v", err)
	}

	return c != 0
}

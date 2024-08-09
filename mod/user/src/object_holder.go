package user

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

var _ objects.Holder = &Module{}

func (mod *Module) HoldObject(id object.ID) (hold bool) {
	mod.db.
		Model(&dbNodeContract{}).
		Where("object_id = ? AND expires_at > ?", id, time.Now().UTC()).
		Select("count(*) > 0").
		First(&hold)

	return
}

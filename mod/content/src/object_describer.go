package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (list []*objects.SourcedObject) {
	var err error
	var row dbDataType

	err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err == nil {
		if row.Method == "adc" {
			list = append(list, &objects.SourcedObject{
				Source: mod.node.Identity(),
				Object: &content.ObjectDescriptor{Type: row.Type},
			})
		}
	}

	return
}

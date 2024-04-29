package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(_ context.Context, objectID object.ID, _ *desc.Opts) []*desc.Desc {
	var descs []*desc.Desc
	var err error
	var row dbDataType

	err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err == nil {
		descs = append(descs, &desc.Desc{
			Source: mod.node.Identity(),
			Data: content.TypeDesc{
				Method:      row.Method,
				ContentType: row.Type,
			},
		})
	}

	if label := mod.GetLabel(objectID); label != "" {
		descs = append(descs, &desc.Desc{
			Source: mod.node.Identity(),
			Data:   content.LabelDesc{Label: label},
		})
	}

	return descs
}

package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	var err error
	var row dbDataType
	var results = make(chan *objects.SourcedObject, 1)
	defer close(results)

	err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	if err != nil {
		return nil, err
	}

	if row.Method == "adc" {
		results <- &objects.SourcedObject{
			Source: mod.node.Identity(),
			Object: &content.ObjectDescriptor{Type: row.Type},
		}
	}

	return results, nil
}

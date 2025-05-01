package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"path/filepath"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	rows, err := mod.db.FindByObjectID(objectID)
	if err != nil {
		return nil, err
	}

	var results = make(chan *objects.SourcedObject, len(rows))
	defer close(results)

	for _, row := range rows {
		base := filepath.Base(row.Path)

		results <- &objects.SourcedObject{
			Source: mod.node.Identity(),
			Object: (*fs.FileName)(&base),
		}
	}

	return results, nil
}

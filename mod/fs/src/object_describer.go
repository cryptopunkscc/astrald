package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	var rows []dbLocalFile

	err := mod.db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Select("path", "mod_time").
		Find(&rows).
		Error
	if err != nil {
		return nil, err
	}

	var results = make(chan *objects.SourcedObject, len(rows))
	defer close(results)

	for _, row := range rows {
		results <- &objects.SourcedObject{
			Source: mod.node.Identity(),
			Object: &fs.FileDescriptor{
				Path:    astral.String16(row.Path),
				ModTime: astral.Time(row.ModTime),
			},
		}
	}

	return results, nil
}

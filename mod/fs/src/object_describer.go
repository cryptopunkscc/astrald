package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

var _ objects.Describer = &Module{}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (list []*objects.SourcedObject) {
	if !scope.Is(astral.ZoneDevice) {
		return
	}

	var rows []dbLocalFile

	err := mod.db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Select("path", "mod_time").
		Find(&rows).
		Error
	if err != nil {
		return
	}

	for _, row := range rows {
		list = append(list, &objects.SourcedObject{
			Source: mod.node.Identity(),
			Object: &fs.FileDescriptor{
				Path:    row.Path,
				ModTime: row.ModTime,
			},
		})
	}

	return
}

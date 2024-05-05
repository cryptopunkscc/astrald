package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) (desc []*desc.Desc) {
	desc = append(desc, mod.describeArchive(objectID)...)
	desc = append(desc, mod.describeMember(objectID)...)

	return
}

func (mod *Module) describeArchive(objectID object.ID) []*desc.Desc {
	var data archives.ArchiveDesc
	var archive = mod.getCache(objectID)
	if archive == nil {
		return nil
	}

	for _, e := range archive.Entries {
		data.Files = append(data.Files, archives.ArchiveEntry{
			ObjectID: e.ObjectID,
			Path:     e.Path,
		})
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   data,
	}}
}

func (mod *Module) describeMember(objectID object.ID) []*desc.Desc {
	var rows []*dbEntry

	mod.db.
		Where("object_id = ?", objectID).
		Preload("Parent").
		Find(&rows)

	if len(rows) == 0 {
		return nil
	}

	var data archives.EntryDesc

	for _, row := range rows {
		data.Containers = append(data.Containers, archives.Container{
			ObjectID: row.Parent.ObjectID,
			Path:     row.Path,
		})
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   data,
	}}
}

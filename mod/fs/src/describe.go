package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/resources"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var fd fs.FileDesc
	var rows []*dbLocalFile

	query := mod.db.Where("data_id = ?", dataID)

	// exclude rw paths
	for _, path := range mod.readwrite.paths.Clone() {
		query = query.Where("path not like ?", path+"/%")
	}

	// exclude astral's root dir
	if fr, ok := mod.assets.Res().(*resources.FileResources); ok {
		query = query.Where("path not like ?", fr.Root()+"/%")
	}

	err := query.Find(&rows).Error
	if err != nil {
		mod.log.Errorv(1, "describe: database error: %v", err)
		return nil
	}
	if len(rows) == 0 {
		return nil
	}

	for _, file := range rows {
		fd.Paths = append(fd.Paths, file.Path)
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   fd,
	}}
}

package fs

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/resources"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []content.Descriptor {
	var desc fs.FileDescriptor
	var rows []*dbLocalFile

	query := mod.db.Where("data_id = ?", dataID)

	// exclude store paths
	for _, path := range mod.store.paths.Clone() {
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
		mod.log.Errorv(1, "describe: no paths")
		return nil
	}

	for _, file := range rows {
		fmt.Println(file.Path)
		desc.Paths = append(desc.Paths, file.Path)
	}

	return []content.Descriptor{desc}
}

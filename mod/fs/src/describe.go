package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []content.Descriptor {
	var desc fs.FileDescriptor
	var files = mod.dbFindByID(dataID)

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		desc.Paths = append(desc.Paths, file.Path)
	}

	return []content.Descriptor{desc}
}

package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	_data "github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
)

func (mod *Module) DescribeData(ctx context.Context, dataID data.ID, opts *_data.DescribeOpts) []_data.Descriptor {
	var desc fs.FileDescriptor
	var files = mod.dbFindByID(dataID)

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		desc.Paths = append(desc.Paths, file.Path)
	}

	return []_data.Descriptor{
		{
			Type: fs.FileDescriptorType,
			Data: desc,
		},
	}
}

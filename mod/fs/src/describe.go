package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/fs"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	paths := mod.path(dataID)

	if len(paths) == 0 {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   fs.FileDesc{Paths: paths},
	}}
}

package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	paths := mod.path(objectID)

	if len(paths) == 0 {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   fs.FileDesc{Paths: paths},
	}}
}

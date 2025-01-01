package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

func (mod *Module) opInfo(ctx astral.Context, q shell.Query) (err error) {
	// authorize
	if !mod.Auth.Authorize(ctx.Identity(), fs.ActionManage, nil) {
		return q.Reject()
	}

	// accept
	stream, err := shell.AcceptStream(q)
	if err != nil {
		return err
	}
	defer stream.Close()

	// prepare data
	paths := mod.watcher.List()
	slices.Sort(paths)

	// output data
	for _, path := range paths {
		_, err = stream.WriteObject((*fs.Path)(&path))
		if err != nil {
			return
		}
	}

	return
}

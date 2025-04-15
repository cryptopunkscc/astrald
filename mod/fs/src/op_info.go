package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opInfoArgs struct {
	Format string `query:"optional"`
}

func (mod *Module) opInfo(ctx *astral.Context, q shell.Query, args opInfoArgs) (err error) {
	// authorize
	if !mod.Auth.Authorize(q.Caller(), fs.ActionManage, nil) {
		return q.Reject()
	}

	// accept the connection
	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	// prepare data
	paths := mod.watcher.List()
	slices.Sort(paths)

	// write results
	for _, path := range paths {
		err = ch.Write((*fs.Path)(&path))
		if err != nil {
			return
		}
	}

	return
}

package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

type opInfoArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpInfo(ctx *astral.Context, q shell.Query, args opInfoArgs) (err error) {
	// accept the connection
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
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

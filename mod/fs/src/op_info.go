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
	if !mod.Auth.Authorize(ctx.Identity(), fs.ActionManage, nil) {
		return q.Reject()
	}

	// accept
	conn := q.Accept()
	defer conn.Close()

	// prepare data
	paths := mod.watcher.List()
	slices.Sort(paths)

	switch args.Format {
	case "json":
		stream := shell.NewJSONStream(conn)
		for _, path := range paths {
			_, err = stream.WriteObject((*fs.Path)(&path))
			if err != nil {
				return
			}
		}

	default:
		stream := astral.NewStream(conn, mod.Objects.Blueprints())
		for _, path := range paths {
			_, err = stream.WriteObject((*fs.Path)(&path))
			if err != nil {
				return
			}
		}
	}

	return
}

package ops

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

func Serve(ctx *astral.Context, set *Set) error {
	srv, err := astrald.Listen()
	if err != nil {
		return err
	}

	return srv.Serve(ctx, set)
}

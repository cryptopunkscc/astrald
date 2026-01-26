package ops

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	apphost "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func Serve(ctx *astral.Context, set *Set) error {
	srv, err := astrald.Listen()
	if err != nil {
		return err
	}

	err = apphost.RegisterHandler(ctx, srv.Endpoint(), srv.AuthToken())
	if err != nil {
		return err
	}

	return srv.Serve(ctx, set)
}

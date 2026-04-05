package ops

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	apphost "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

func Serve(ctx *astral.Context, set *Set, opts ...apphost.RegistrarOption) error {
	reg := apphost.NewRegistrar(apphost.Default(), opts...)
	srv, err := astrald.NewHandler(ctx, reg)
	if err != nil {
		return err
	}
	return srv.Route(ctx, set)
}

package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	jrpc "github.com/cryptopunkscc/go-apphost-jrpc"
	"log"
)

func RunIdentityFilterService(ctx context.Context, filter id.Filter) (port string, err error) {
	var portId id.Identity
	if portId, err = id.GenerateIdentity(); err != nil {
		return
	}
	port = portId.String()
	srv := jrpc.NewApp(port)
	srv.Logger(log.New(log.Writer(), "filter service ", 0))
	srv.Func("", func(_, identity id.Identity) bool {
		return filter(identity)
	})
	err = srv.Run(ctx)
	return
}

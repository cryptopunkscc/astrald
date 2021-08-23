package share

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const (
	Port = "shared"
	RemotePort = "shares"
)

const (
	Add      = 1
	Remove   = 2
	List     = 3
	Contains = 4
)

func (sc serviceContext) runLocal(ctx context.Context, core api.Core) error {
	rc := requestContext{sc}
	handlers := map[byte]request.Handler{
		Add:      rc.Add,
		Remove:   rc.Remove,
		List:     rc.ListLocal,
		Contains: rc.ContainsLocal,
	}
	handle.Requests(ctx, core, Port, auth.Local, handle.Using(handlers))
	return nil
}

func (sc serviceContext) runRemote(ctx context.Context, core api.Core) error {
	rc := requestContext{sc}
	handlers := map[byte]request.Handler{
		List:     rc.List,
		Contains: rc.Contains,
	}
	handle.Requests(ctx, core, RemotePort, auth.All, handle.Using(handlers))
	return nil
}

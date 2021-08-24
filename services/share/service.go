package share

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/shares"
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

type service struct {
	shared shares.Shared
}

func (srv *service) runLocal(ctx context.Context, core api.Core) error {
	handlers := request.Handlers{
		Add:      srv.Add,
		Remove:   srv.Remove,
		List:     srv.ListLocal,
		Contains: srv.ContainsLocal,
	}
	handle.Requests(ctx, core, Port, auth.Local, handle.Using(handlers))
	return nil
}

func (srv *service) runRemote(ctx context.Context, core api.Core) error {
	handlers := request.Handlers{
		List:     srv.List,
		Contains: srv.Contains,
	}
	handle.Requests(ctx, core, RemotePort, auth.All, handle.Using(handlers))
	return nil
}

package identity

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/components/uid"
	"github.com/cryptopunkscc/astrald/components/uid/file"
	"github.com/cryptopunkscc/astrald/services"
	"github.com/cryptopunkscc/astrald/services/util/auth"
	"github.com/cryptopunkscc/astrald/services/util/handle"
	"github.com/cryptopunkscc/astrald/services/util/request"
)

const (
	Update = 1
	List   = 2
	Get    = 3
)

const Port = "id"

type service struct {
	ids uid.Identities
}

func (srv *service) runService(ctx context.Context, core api.Core) error {
	srv.ids = file.NewIdentities(services.AstralHome)
	handlers := request.Handlers{
		Update: srv.Update,
		List:   srv.List,
		Get: srv.Get,
	}
	handle.Requests(ctx, core, Port, auth.Local, handle.Using(handlers))
	return nil
}

package identity

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
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

func (c *Context) runService(ctx context.Context, core api.Core) error {
	c.ids = file.NewIdentities(services.AstralHome)
	rc := Request{c}
	handlers := request.Handlers{
		Update: rc.Update,
		List:   rc.List,
		Get: rc.Get,
	}
	handle.Requests(ctx, core, Port, auth.Local, handle.Using(handlers))
	return nil
}

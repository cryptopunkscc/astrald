package astrald

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/apphost"
)

type Router interface {
	RouteQuery(*astral.Context, *astral.Query) (astral.Conn, error)
	GuestID() *astral.Identity
	HostID() *astral.Identity
}

var _ Router = &apphost.Router{}

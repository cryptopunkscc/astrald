package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type Route struct {
	Caller   id.Identity
	Target   id.Identity
	Router   net.Router
	Priority int
}

func MatchRoutes(routes []Route, caller id.Identity, target id.Identity) []Route {
	var matched []Route

	for _, route := range routes {
		if (route.Caller.IsEqual(id.Anyone) || route.Caller.IsEqual(caller)) &&
			(route.Target.IsEqual(id.Anyone) || route.Target.IsEqual(target)) {
			matched = append(matched, route)
		}
	}

	return matched
}

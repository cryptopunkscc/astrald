package core

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node"
)

func MatchRoutes(routes []node.Route, caller id.Identity, target id.Identity) []node.Route {
	var matched []node.Route

	for _, route := range routes {
		if (route.Caller.IsEqual(id.Anyone) || route.Caller.IsEqual(caller)) &&
			(route.Target.IsEqual(id.Anyone) || route.Target.IsEqual(target)) {
			matched = append(matched, route)
		}
	}

	return matched
}

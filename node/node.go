package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/authorizer"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
	"regexp"
)

type Node interface {
	Identity() id.Identity
	Events() *events.Queue
	Infra() infra.Infra
	Network() network.Network
	Auth() authorizer.AuthSet
	Modules() modules.Modules
	Resolver() resolver.ResolveEngine
	Router() router.Router
	LocalRouter() router.LocalRouter
}

// FormatString replaces public keys in the string with identity aliases. Public keys must
// be surrounded by double curly braces {{ and }}.
func FormatString(node Node, s string) string {
	pattern := regexp.MustCompile(`\{\{([0-9A-Za-z:_-]*)}}`)

	// Replace matches using a custom function
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the string inside the curly braces
		val := pattern.FindStringSubmatch(match)[1]

		for _, f := range formatters {
			if s := f(node, val); s != "" {
				return s
			}
		}

		return match
	})
}

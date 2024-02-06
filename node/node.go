package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"regexp"
)

type Node interface {
	Identity() id.Identity
	Events() *events.Queue
	Infra() infra.Infra
	Network() network.Network
	Tracker() tracker.Tracker
	Modules() modules.Modules
	Resolver() resolver.Resolver
	Router() router.Router
	LocalRouter() router.LocalRouter
}

// FormatString replaces public keys in the string with identity aliases. Public keys must
// be surrounded by double curly braces {{ and }}.
func FormatString(node Node, s string) string {
	pattern := regexp.MustCompile(`\{\{([0-9A-Fa-f]{66})}}`)

	// Replace matches using a custom function
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract the hex string inside the curly braces
		hex := pattern.FindStringSubmatch(match)[1]

		identity, err := id.ParsePublicKeyHex(hex)
		if err != nil {
			return match
		}

		name := node.Resolver().DisplayName(identity)

		// Return the modified string
		return name
	})
}

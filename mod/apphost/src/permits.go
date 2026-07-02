package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

// permitsForOrigin returns the permit template granted to an identity that
// registers from the given web origin, or nil when the origin is not recognized.
// The returned set is the capability template for that origin (e.g. the
// "astrald.settings" template); it grows as settings operations become
// permit-guarded.
func (mod *Module) permitsForOrigin(origin string) []*auth.Permit {
	switch origin {
	case "https://settings.astrald.app":
		return []*auth.Permit{
			// note: baseline capability; extend with settings-specific action
			// permits as those ops start consulting Auth.Authorize.
			{Action: astral.String8(nodes.RelayForAction{}.ObjectType())},
		}
	default:
		return nil
	}
}

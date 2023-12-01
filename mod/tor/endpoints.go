package tor

import "github.com/cryptopunkscc/astrald/net"

func (mod *Module) Endpoints() []net.Endpoint {
	if (mod.server == nil) || mod.server.endpoint.IsZero() {
		return nil
	}
	return []net.Endpoint{mod.server.endpoint}
}

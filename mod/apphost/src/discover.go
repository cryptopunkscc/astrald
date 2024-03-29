package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
)

const SourceServiceName = "discovery.source"

func (mod *Module) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	mod.guestsMu.Lock()
	var guests []*Guest
	for _, guest := range mod.guests {
		guests = append(guests, guest)
	}
	mod.guestsMu.Unlock()

	var list = make([]discovery.Service, 0)

	for _, guest := range guests {
		var query = net.NewQuery(caller, guest.Identity, SourceServiceName)

		conn, err := net.Route(ctx, guest.router, query)
		if err != nil {
			continue
		}

		for {
			var entry discovery.Service
			if err := cslq.Decode(conn, "v", &entry); err != nil {
				break
			}
			// force identity to match the service
			entry.Identity = guest.Identity
			list = append(list, entry)
		}
		conn.Close()
	}

	return list, nil
}

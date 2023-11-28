package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	mod.guestsMu.Lock()
	var guests []*Guest
	for _, guest := range mod.guests {
		guests = append(guests, guest)
	}
	mod.guestsMu.Unlock()

	var list = make([]sdp.ServiceEntry, 0)

	for _, guest := range guests {
		var query = net.NewQuery(caller, guest.Identity, sdp.SourceServiceName)

		conn, err := net.Route(ctx, guest.router, query)
		if err != nil {
			continue
		}

		for {
			var entry sdp.ServiceEntry
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

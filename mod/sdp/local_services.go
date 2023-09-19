package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/services"
)

const SourceServiceName = "core.sdp.source"

type LocalServices struct {
	services services.Services
}

func NewLocalServices(services services.Services) *LocalServices {
	return &LocalServices{services: services}
}

func (l *LocalServices) Discover(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error) {
	var list = make([]ServiceEntry, 0)

	for _, service := range l.services.FindByName(SourceServiceName) {
		query := net.NewQuery(caller, service.Identity(), SourceServiceName)
		conn, err := net.Route(ctx, l.services, query)
		if err != nil {
			continue
		}

		for {
			var entry ServiceEntry
			if err := cslq.Decode(conn, "v", &entry); err != nil {
				break
			}
			// force identity to match the service
			entry.Identity = service.Identity()
			list = append(list, entry)
		}
		conn.Close()
	}

	return list, nil
}

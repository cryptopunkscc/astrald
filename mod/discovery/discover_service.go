package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/query"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

var _ tasks.Runner = &DiscoveryService{}

type DiscoveryService struct {
	*Module
}

const discoverServiceName = "services.discover"

func (service *DiscoveryService) RouteQuery(ctx context.Context, q query.Query, swc net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	return query.Accept(q, swc, func(conn net.SecureConn) {
		service.serveDiscover(conn, q.Origin())
	})
}

func (service *DiscoveryService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), discoverServiceName, service)
	if err != nil {
		return err
	}

	<-s.Done()

	return nil
}

func (service *DiscoveryService) serveDiscover(conn net.SecureConn, origin string) {
	var session = rpc.New(conn)

	var wg sync.WaitGroup

	for source, sourceID := range service.sources {
		source, sourceID := source, sourceID

		wg.Add(1)
		go func() {
			defer wg.Done()

			list, err := source.Discover(service.ctx, conn.RemoteIdentity(), origin)
			if err != nil {
				return
			}

			for _, item := range list {
				item.Identity = sourceID
				session.Encode(item)
			}
		}()
	}

	wg.Wait()
}

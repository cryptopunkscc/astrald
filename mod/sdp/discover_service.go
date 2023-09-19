package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

var _ tasks.Runner = &DiscoveryService{}

type DiscoveryService struct {
	*Module
}

const DiscoverServiceName = "core.sdp.discover"

func (service *DiscoveryService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), DiscoverServiceName, service)
	if err != nil {
		return err
	}

	<-s.Done()

	return nil
}

func (service *DiscoveryService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer debug.SaveLog(func(p any) {
			service.log.Error("discovery panicked: %v", p)
		})

		service.serve(conn, hints.Origin)
	})
}

func (service *DiscoveryService) serve(conn net.SecureConn, origin string) {
	defer conn.Close()

	var wg sync.WaitGroup

	service.log.Logv(1, "discovery request from %v (%s)", conn.RemoteIdentity(), origin)

	for source := range service.sources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			list, err := source.Discover(service.ctx, conn.RemoteIdentity(), origin)
			if err != nil {
				return
			}

			for _, item := range list {
				cslq.Encode(conn, "v", item)
			}
		}()
	}

	wg.Wait()
}

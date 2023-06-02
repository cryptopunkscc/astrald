package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

var _ tasks.Runner = &DiscoveryService{}

type DiscoveryService struct {
	*Module
}

const discoverServiceName = "services.discover"

func (m *DiscoveryService) Run(ctx context.Context) error {
	var workers = RunQueryWorkers(ctx, m.handleQuery, 8)

	service, err := m.node.Services().Register(ctx, m.node.Identity(), discoverServiceName, workers.Enqueue)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil
}

func (m *DiscoveryService) handleQuery(ctx context.Context, query *services.Query) error {
	nconn, err := query.Accept()
	if err != nil {
		return err
	}
	defer nconn.Close()

	var conn = proto.New(nconn)

	var wg sync.WaitGroup

	for source := range m.sources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			list, err := source.Discover(ctx, query.RemoteIdentity(), query.Origin())
			if err != nil {
				return
			}

			for _, item := range list {
				conn.Encode(item)
			}
		}()
	}

	wg.Wait()

	return nil
}

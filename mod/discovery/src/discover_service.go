package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &DiscoveryService{}

type DiscoveryService struct {
	*Module
}

func (service *DiscoveryService) Run(ctx context.Context) error {
	err := service.AddRoute(discovery.DiscoverServiceName, service)
	if err != nil {
		return err
	}
	defer service.RemoveRoute(discovery.DiscoverServiceName)

	<-ctx.Done()

	return nil
}

func (service *DiscoveryService) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer debug.SaveLog(func(p any) {
			service.log.Error("discovery panicked: %v", p)
		})

		var err = service.serve(conn, hints.Origin)
		if err != nil {
			service.log.Errorv(1, "error serving %v: %v", caller.Identity(), err)
		}
	})
}

func (service *DiscoveryService) serve(conn astral.Conn, origin string) error {
	defer conn.Close()

	service.log.Logv(1, "discovery request from %v (%s)", conn.RemoteIdentity(), origin)

	info, err := service.DiscoverLocal(service.ctx, conn.RemoteIdentity(), origin)
	if err != nil {
		return err
	}

	var protoInfo proto.Info
	for _, c := range info.Data {
		protoInfo.Data = append(protoInfo.Data, proto.Data{Bytes: c.Bytes})
	}

	for _, s := range info.Services {
		protoInfo.Services = append(protoInfo.Services, proto.Service{
			Identity: s.Identity,
			Name:     s.Name,
			Type:     s.Type,
			Extra:    s.Extra,
		})
	}

	return cslq.Encode(conn, "v", &protoInfo)
}

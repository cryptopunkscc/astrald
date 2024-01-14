package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ tasks.Runner = &DiscoveryService{}

type DiscoveryService struct {
	*Module
}

func (service *DiscoveryService) Run(ctx context.Context) error {
	err := service.node.LocalRouter().AddRoute(discovery.DiscoverServiceName, service)
	if err != nil {
		return err
	}
	defer service.node.LocalRouter().RemoveRoute(discovery.DiscoverServiceName)

	<-ctx.Done()

	return nil
}

func (service *DiscoveryService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer debug.SaveLog(func(p any) {
			service.log.Error("discovery panicked: %v", p)
		})

		var err = service.serve(conn, hints.Origin)
		if err != nil {
			service.log.Errorv(1, "error serving %v: %v", caller.Identity(), err)
		}
	})
}

func (service *DiscoveryService) serve(conn net.SecureConn, origin string) error {
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

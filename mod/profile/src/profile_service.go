package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/astral"
	"time"
)

type ProfileService struct {
	*Module
}

func (service *ProfileService) Run(ctx context.Context) error {
	var err = service.AddRoute(serviceName, service)
	if err != nil {
		return err
	}
	defer service.RemoveRoute(serviceName)

	if service.sdp != nil {
		service.sdp.AddServiceDiscoverer(service)
		defer service.sdp.RemoveServiceDiscoverer(service)
	}

	<-ctx.Done()
	return err
}

func (service *ProfileService) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	return []discovery.Service{
		{
			Identity: service.node.Identity(),
			Name:     serviceName,
			Type:     serviceType,
			Extra:    nil,
		},
	}, nil
}

func (service *ProfileService) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		service.serve(conn)
	})
}

func (service *ProfileService) serve(conn astral.Conn) {
	defer conn.Close()

	service.log.Infov(2, "%s asked for profile", conn.RemoteIdentity())

	json.NewEncoder(conn).Encode(service.getLocalProfile())
}

func (service *ProfileService) getLocalProfile() *proto.Profile {
	var p = &proto.Profile{
		Endpoints: []proto.Endpoint{},
	}

	p.Alias = service.dir.DisplayName(service.node.Identity())

	endpoints, _ := service.exonet.Resolve(context.Background(), service.node.Identity())

	for _, a := range endpoints {
		p.Endpoints = append(p.Endpoints, proto.Endpoint{
			Network:   a.Network(),
			Address:   a.Address(),
			Public:    false,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365),
		})
	}

	return p
}

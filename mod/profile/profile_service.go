package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

type ProfileService struct {
	*Module
}

func (service *ProfileService) Run(ctx context.Context) error {
	s, err := service.node.Services().Register(ctx, service.node.Identity(), serviceName, service)

	disco, err := modules.Find[*discovery.Module](service.node.Modules())
	if err == nil {
		disco.AddSourceContext(ctx, service, service.node.Identity())
	} else {
		service.log.Errorv(2, "can't regsiter service discovery source: %s", err)
	}

	<-s.Done()
	return err
}

func (service *ProfileService) Discover(ctx context.Context, caller id.Identity, origin string) ([]discovery.ServiceEntry, error) {
	return []discovery.ServiceEntry{
		{
			Identity: service.node.Identity(),
			Name:     serviceName,
			Type:     serviceType,
			Extra:    nil,
		},
	}, nil
}

func (service *ProfileService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		service.serve(conn)
	})
}

func (service *ProfileService) serve(conn net.SecureConn) {
	defer conn.Close()

	service.log.Infov(2, "%s asked for profile", conn.RemoteIdentity())

	json.NewEncoder(conn).Encode(service.getLocalProfile())
}

func (service *ProfileService) getLocalProfile() *proto.Profile {
	var p = &proto.Profile{
		Endpoints: []proto.Endpoint{},
	}

	p.Alias = service.node.Resolver().DisplayName(service.node.Identity())

	for _, a := range service.node.Infra().Endpoints() {
		p.Endpoints = append(p.Endpoints, proto.Endpoint{
			Network:   a.Network(),
			Address:   a.String(),
			Public:    false,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365),
		})
	}

	return p
}

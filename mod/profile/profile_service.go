package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type ProfileService struct {
	*Module
}

func (service *ProfileService) Run(ctx context.Context) error {
	var err = service.node.LocalRouter().AddRoute(serviceName, service)
	if err != nil {
		return err
	}
	defer service.node.LocalRouter().RemoveRoute(serviceName)

	if service.sdp != nil {
		service.sdp.AddSource(service)
		defer service.sdp.RemoveSource(service)
	}

	<-ctx.Done()
	return err
}

func (service *ProfileService) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	return []sdp.ServiceEntry{
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

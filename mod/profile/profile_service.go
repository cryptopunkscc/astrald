package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

type ProfileService struct {
	*Module
}

func (service *ProfileService) Run(ctx context.Context) error {
	var err = service.node.AddRoute(serviceName, service)
	if err != nil {
		return err
	}
	defer service.node.RemoveRoute(serviceName)

	disco, err := modules.Find[*sdp.Module](service.node.Modules())
	if err == nil {
		disco.AddSource(service)
		defer disco.RemoveSource(service)
	} else {
		service.log.Errorv(2, "can't regsiter service discovery source: %s", err)
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

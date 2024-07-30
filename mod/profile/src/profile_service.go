package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"io"
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

	<-ctx.Done()
	return err
}

func (service *ProfileService) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (io.WriteCloser, error) {
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

	p.Alias = service.Dir.DisplayName(service.node.Identity())

	endpoints, _ := service.Exonet.Resolve(context.Background(), service.node.Identity())

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

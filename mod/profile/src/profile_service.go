package profile

import (
	"encoding/json"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
)

// TODO: refactor module

type ProfileService struct {
	*Module
}

func (service *ProfileService) Run(ctx *astral.Context) error {
	var err = service.AddRoute(serviceName, service)
	if err != nil {
		return err
	}
	defer service.RemoveRoute(serviceName)

	<-ctx.Done()
	return err
}

func (service *ProfileService) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return query.Accept(q, w, func(conn astral.Conn) {
		service.serve(conn)
	})
}

func (service *ProfileService) serve(conn astral.Conn) {
	defer conn.Close()

	service.log.Infov(2, "%v asked for profile", conn.RemoteIdentity())

	json.NewEncoder(conn).Encode(service.getLocalProfile())
}

func (service *ProfileService) getLocalProfile() *proto.Profile {
	var p = &proto.Profile{
		Endpoints:   []proto.Endpoint{},
		HasContract: service.User.ActiveContract() != nil,
	}

	p.Alias = service.Dir.DisplayName(service.node.Identity())

	endpoints, _ := service.Nodes.ResolveEndpoints(astral.NewContext(nil), service.node.Identity())

	for a := range endpoints {
		p.Endpoints = append(p.Endpoints, proto.Endpoint{
			Network:   a.Network(),
			Address:   a.Address(),
			Public:    false,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365),
		})
	}

	return p
}

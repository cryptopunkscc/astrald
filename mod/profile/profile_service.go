package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	discoproto "github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
	"time"
)

type ProfileService struct {
	*Module
}

func (srv *ProfileService) Discover(ctx context.Context, caller id.Identity, medium string) ([]discoproto.ServiceEntry, error) {
	return []discoproto.ServiceEntry{
		{
			Name:  serviceName,
			Type:  serviceType,
			Extra: nil,
		},
	}, nil
}

func (srv *ProfileService) Run(ctx context.Context) error {
	_, err := srv.node.Services().Register(ctx, srv.node.Identity(), serviceName, srv.handle)

	disco, err := modules.Find[*discovery.Module](srv.node.Modules())
	if err == nil {
		disco.AddSource(srv, srv.node.Identity())
		go func() {
			<-ctx.Done()
			disco.RemoveSource(srv)
		}()
	} else {
		srv.log.Errorv(2, "can't regsiter service discovery source: %s", err)
	}

	<-ctx.Done()
	return err
}

func (srv *ProfileService) handle(ctx context.Context, query *services.Query) error {
	conn, err := query.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	srv.log.Infov(2, "%s asked for profile", query.RemoteIdentity())

	return json.NewEncoder(conn).Encode(srv.getLocalProfile())
}

func (srv *ProfileService) getLocalProfile() *proto.Profile {
	var p = &proto.Profile{
		Endpoints: []proto.Endpoint{},
	}

	p.Alias = srv.node.Alias()

	for _, a := range srv.node.Infra().Endpoints() {
		p.Endpoints = append(p.Endpoints, proto.Endpoint{
			Network:   a.Network(),
			Address:   a.String(),
			Public:    false,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365),
		})
	}

	return p
}

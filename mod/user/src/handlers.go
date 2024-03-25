package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) DiscoverData(ctx context.Context, caller id.Identity, origin string) ([][]byte, error) {
	var data [][]byte

	if mod.localUser == nil {
		return nil, nil
	}

	if len(mod.localUser.cert) == 0 {
		return nil, nil
	}

	data = append(data, mod.localUser.cert)

	return data, nil
}

func (mod *Module) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	var services []discovery.Service

	var user = mod.LocalUser()
	if user == nil {
		return nil, nil
	}

	services = append(services,
		discovery.Service{
			Identity: user.Identity(),
			Name:     userProfileServiceName,
			Type:     userProfileServiceType,
			Extra:    nil,
		})

	return services, nil
}

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	user := mod.LocalUser()
	if user == nil {
		return net.RouteNotFound(mod)
	}

	targetID := query.Target()
	if targetID.IsZero() {
		return net.RouteNotFound(mod)
	}

	if !targetID.IsEqual(user.Identity()) {
		return net.RouteNotFound(mod)
	}

	return mod.routes.RouteQuery(ctx, query, caller, hints)
}

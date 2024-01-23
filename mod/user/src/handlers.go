package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) Authorize(identity id.Identity, dataID _data.ID) error {
	_, found := mod.identities.Get(identity.PublicKeyHex())
	if found {
		return nil
	}

	return shares.ErrDenied
}

func (mod *Module) DiscoverData(ctx context.Context, caller id.Identity, origin string) ([][]byte, error) {
	var data [][]byte

	for _, i := range mod.identities.Clone() {
		if len(i.cert) == 0 {
			continue
		}

		data = append(data, i.cert)
	}

	return data, nil
}

func (mod *Module) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	var services []discovery.Service

	for _, i := range mod.identities.Clone() {
		services = append(services,
			discovery.Service{
				Identity: i.identity,
				Name:     userProfileServiceName,
				Type:     userProfileServiceType,
				Extra:    nil,
			})
	}

	return services, nil
}

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	targetID := query.Target()
	if targetID.IsZero() {
		return net.RouteNotFound(mod)
	}

	identity, ok := mod.identities.Get(targetID.PublicKeyHex())
	if !ok {
		return net.RouteNotFound(mod)
	}

	//TODO: find a better way to handle this
	if query.Query() == userProfileServiceName {
		return mod.profileHandler.RouteQuery(ctx, query, caller, hints)
	}

	return identity.routes.RouteQuery(ctx, query, caller, hints)
}

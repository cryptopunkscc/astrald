package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) DiscoverData(ctx context.Context, caller id.Identity, origin string) ([][]byte, error) {
	var data [][]byte

	if mod.UserID().IsZero() {
		return nil, nil
	}

	if len(mod.userCert) == 0 {
		return nil, nil
	}

	data = append(data, mod.userCert)

	return data, nil
}

func (mod *Module) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	var userID = mod.UserID()
	if userID.IsZero() {
		return nil, nil
	}

	return []discovery.Service{{
		Identity: userID,
		Name:     userProfileServiceName,
		Type:     userProfileServiceType,
		Extra:    nil,
	}}, nil
}

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if !query.Target().IsEqual(mod.UserID()) {
		return net.RouteNotFound(mod)
	}

	return mod.routes.RouteQuery(ctx, query, caller, hints)
}

func (mod *Module) discoverUsers(ctx context.Context) {
	events.Handle(ctx, mod.node.Events(), func(event discovery.EventDiscovered) error {
		// make sure we're not getting our own services
		if event.Identity.IsEqual(mod.node.Identity()) {
			return nil
		}

		for _, cert := range event.Info.Data {
			err := mod.checkCert(event.Identity, cert.Bytes)
			if err != nil {
				mod.log.Errorv(2, "checkCert %v from %v: %v", object.Resolve(cert.Bytes), event.Identity, err)
			}
		}

		for _, service := range event.Info.Services {
			// look only for user profiles
			if service.Type != userProfileServiceType {
				continue
			}

			// and only if they're hosted
			if service.Identity.IsEqual(event.Identity) {
				continue
			}

			mod.log.Infov(2, "user %v discovered on %v", service.Identity, event.Identity)

			if len(service.Extra) == 0 {
				continue
			}
		}

		return nil
	})
}

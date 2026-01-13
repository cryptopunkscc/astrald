package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/sig"
)

func (mod *Module) DiscoverService(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (<-chan services.ServiceDiscoveryResult, error) {
	var snapshot []services.ServiceDiscoveryResult
	if opts.Snapshot {
		svc := services.Service{
			Name:        nat.ModuleName,
			Identity:    mod.node.Identity(),
			Composition: astral.NewBundle(),
		}

		change := services.ServiceChange{
			Enabled: astral.Bool(mod.serviceEnabled),
			Service: svc,
		}

		snapshot = append(snapshot, services.DiscoveryChange(change))
	}

	var updates <-chan services.ServiceDiscoveryResult
	if opts.Follow {
		feedUpdates := mod.serviceChangeFeed.Subscribe(ctx)
		updates = sig.MapChan(
			ctx,
			feedUpdates,
			services.DiscoveryChange,
		)
	}

	return sig.SnapshotFollowStream(
		ctx,
		snapshot,
		updates,
		services.DiscoveryFlush(),
	), nil
}

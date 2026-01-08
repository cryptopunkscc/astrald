package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
)

const ModuleName = "services"

type Module struct {
	Deps

	node astral.Node
	log  *log.Logger
	ops  shell.Scope
	db   *DB

	discoverers sig.Set[services.ServiceDiscoverer]
}

var _ services.Module = &Module{}

func (mod *Module) AddServiceDiscoverer(discoverer services.ServiceDiscoverer) error {
	return mod.discoverers.Add(discoverer)
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

func (mod *Module) String() string {
	return ModuleName
}

func (mod *Module) DiscoverRemoteServices(ctx *astral.Context, target *astral.Identity, subscribe bool) error {
	if target.IsEqual(mod.node.Identity()) {
		return nil
	}

	mod.log.Info("discovering services from %v", target)
	ch, err := query.RouteChan(
		ctx.IncludeZone(astral.ZoneNetwork),
		mod.node,
		query.New(ctx.Identity(), target, services.MethodServiceDiscovery, nil),
	)
	if err != nil {
		return err
	}
	defer ch.Close()

	// Track snapshot state across the whole stream.
	snapshotCompleted := false
	snapshot := make([]services.ServiceChange, 0)

	err = ch.Collect(func(object astral.Object) error {
		switch m := object.(type) {
		case *astral.EOS:
			// End-of-snapshot marker.

			// Replace services from this identity atomically based on collected snapshot.
			var servicesList []services.Service
			for _, svcChange := range snapshot {
				if svcChange.Enabled {
					servicesList = append(servicesList, svcChange.Service)
				}
			}

			err := mod.db.InTx(func(tx *DB) error {
				if err := tx.RemoveIdentityServices(target); err != nil {
					return err
				}

				for _, svc := range servicesList {
					if err := tx.InsertService(&svc); err != nil {
						return err
					}
				}

				return nil
			})
			if err != nil {
				mod.log.Error("failed to replace services from %v: %v", target, err)
				return err
			}

			snapshotCompleted = true

			if !subscribe {
				return nil
			}

		case *services.ServiceChange:
			if !m.Service.Identity.IsEqual(target) {
				mod.log.Info("ignoring service from different identity: %v", m.Service.Identity)
				return nil
			}

			if !snapshotCompleted {
				// Still collecting snapshot.
				snapshot = append(snapshot, *m)
				return nil
			}

			// Live update after snapshot completed.
			err = mod.db.InTx(func(tx *DB) error {
				if err := tx.RemoveIdentityService(m.Service.Name, target); err != nil {
					return err
				}

				if m.Enabled {
					return tx.InsertService(&m.Service)
				}

				return nil
			})
			if err != nil {
				mod.log.Error("failed to update service %v from %v: %v", m.Service.Name, target, err)
			}

		default:
			mod.log.Info("unexpected message type: %T", object)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

package services

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

const ModuleName = "services"

type Module struct {
	Deps

	node astral.Node
	log  *log.Logger
	ops  shell.Scope
	db   *DB

	discoverers []services.ServiceDiscoverer
}

var _ services.Module = &Module{}

func (mod *Module) AddServiceDiscoverer(discoverer services.ServiceDiscoverer) {
	mod.discoverers = append(mod.discoverers, discoverer)
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

	defaultExpiration := astral.Duration(7 * 24 * time.Hour)
	for {
		msg, err := ch.Read()
		if err != nil {
			return err
		}

		var snapshotCompleted = false
		var snapshot = make([]services.ServiceChange, 0)

		switch m := msg.(type) {
		case *astral.EOS:
			// FIXME: in tx
			err := mod.db.InvalidateServices(target)
			if err != nil {
				mod.log.Error("failed to invalidate services from %v: %v", target, err)
			}

			for _, svcChange := range snapshot {
				err := mod.db.InsertService(svcChange.Service, defaultExpiration)
				if err != nil {
					mod.log.Error("failed to insert service %v from %v: %v", svcChange.Service.Name, target, err)
				}
			}

			snapshotCompleted = true

			if !subscribe {
				return nil
			}
		case *services.ServiceChange:
			if !m.Service.Identity.IsEqual(target) {
				mod.log.Info("ignoring service from different identity: %v", m.Service.Identity)
				continue
			}

			if !snapshotCompleted {
				snapshot = append(snapshot, *m)
			}

			if snapshotCompleted {
				err = mod.db.InvalidateService(m.Service.Name, target)
				if err != nil {
					mod.log.Error("failed to invalidate service %v from %v: %v", m.Service.Name, target, err)
				}

				if m.Enabled {
					err := mod.db.InsertService(m.Service, defaultExpiration)
					if err != nil {
						mod.log.Error("failed to insert service %v from %v: %v", m.Service.Name, target, err)
					}
				}
			}
		default:
			mod.log.Info("unexpected message type: %T", msg)
		}
	}
}

package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/services"
	servicescli "github.com/cryptopunkscc/astrald/mod/services/client"
	"github.com/cryptopunkscc/astrald/sig"
)

const ModuleName = "services"

type Module struct {
	Deps

	node astral.Node
	log  *log.Logger
	ops  ops.Set
	db   *DB

	discoverers sig.Set[services.Discoverer]
}

var _ services.Module = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) syncServices(ctx *astral.Context, providerID *astral.Identity, follow bool) error {
	client := servicescli.New(providerID, nil)

	ch, err := client.Discover(ctx, follow)
	if err != nil {
		return err
	}

	// clear cache
	err = mod.db.deleteAllProviderServices(providerID)
	if err != nil {
		return err
	}

	// process updates
	for update := range ch {
		switch {
		case update == nil:
			continue
		case update.Available:
			err = mod.db.createProviderService(update.ProviderID, string(update.Name), update.Info)
		default:
			err = mod.db.deleteProviderService(update.ProviderID, string(update.Name))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (mod *Module) AddDiscoverer(discoverer services.Discoverer) error {
	return mod.discoverers.Add(discoverer)
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return ModuleName
}

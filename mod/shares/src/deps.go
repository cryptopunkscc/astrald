package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
	"strings"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.storage, err = modules.Load[storage.Module](mod.node, storage.ModuleName)
	if err != nil {
		return err
	}

	mod.index, err = modules.Load[index.Module](mod.node, index.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(shares.ModuleName, NewAdmin(mod))
	}

	mod.index.CreateIndex(publicIndexName, index.TypeSet)
	mod.storage.AddReader(shares.ModuleName, mod)

	go events.Handle(context.Background(), mod.node.Events(),
		func(ctx context.Context, event index.EventEntryUpdate) error {
			_, s, found := strings.Cut(event.IndexName, localShareIndexPrefix+".")
			if !found {
				return nil
			}

			identity, err := id.ParsePublicKeyHex(s)
			if err != nil {
				return nil
			}

			mod.Notify(identity)
			return nil
		})

	return err
}

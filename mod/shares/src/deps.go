package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/sets"
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

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(shares.ModuleName, NewAdmin(mod))
	}

	mod.sets.SetOpener(shares.SetType, mod.setOpener)
	mod.sets.Create(publicSetName, sets.TypeBasic)
	mod.storage.AddReader(shares.ModuleName, mod)

	mod.remoteShares, err = sets.Open[sets.Union](mod.sets, shares.RemoteSharesSetName)
	if err != nil {
		mod.remoteShares, err = mod.sets.CreateUnion(shares.RemoteSharesSetName)
		if err != nil {
			return err
		}
		mod.sets.SetDescription(shares.RemoteSharesSetName, "All data from remote shares")
		mod.sets.SetVisible(shares.RemoteSharesSetName, true)
		mod.sets.Universe().Add(shares.RemoteSharesSetName)
	}

	go events.Handle(context.Background(), mod.node.Events(),
		func(event sets.EventMemberUpdate) error {
			_, s, found := strings.Cut(event.Set, localShareSetPrefix+".")
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

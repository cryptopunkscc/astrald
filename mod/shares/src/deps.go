package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
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

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(shares.ModuleName, NewAdmin(mod))
	}

	mod.sets.SetWrapper(shares.RemoteSetType, mod.remoteShareWrapper)
	mod.storage.AddOpener(shares.ModuleName, mod, 10)
	mod.content.AddDescriber(mod)

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

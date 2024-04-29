package user

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// load required dependencies
	mod.objects, err = modules.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.shares, err = modules.Load[shares.Module](mod.node, shares.ModuleName)
	if err != nil {
		return err
	}

	mod.relay, err = modules.Load[relay.Module](mod.node, relay.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = modules.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = modules.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	// load optional dependencies
	mod.content, _ = modules.Load[content.Module](mod.node, content.ModuleName)
	mod.sdp, _ = modules.Load[discovery.Module](mod.node, discovery.ModuleName)
	mod.keys, _ = modules.Load[keys.Module](mod.node, keys.ModuleName)
	mod.admin, _ = modules.Load[admin.Module](mod.node, admin.ModuleName)
	mod.apphost, _ = modules.Load[apphost.Module](mod.node, apphost.ModuleName)

	if mod.sdp != nil {
		mod.sdp.AddServiceDiscoverer(mod)
		mod.sdp.AddDataDiscoverer(mod)
	}

	mod.dir.AddDescriber(mod)

	return nil
}

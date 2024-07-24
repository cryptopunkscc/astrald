package user

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// load required dependencies
	mod.auth, err = core.Load[auth.Module](mod.node, auth.ModuleName)
	if err != nil {
		return err
	}

	mod.objects, err = core.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.shares, err = core.Load[shares.Module](mod.node, shares.ModuleName)
	if err != nil {
		return err
	}

	mod.relay, err = core.Load[relay.Module](mod.node, relay.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = core.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	// load optional dependencies
	mod.content, _ = core.Load[content.Module](mod.node, content.ModuleName)
	mod.sdp, _ = core.Load[discovery.Module](mod.node, discovery.ModuleName)
	mod.keys, _ = core.Load[keys.Module](mod.node, keys.ModuleName)
	mod.admin, _ = core.Load[admin.Module](mod.node, admin.ModuleName)
	mod.apphost, _ = core.Load[apphost.Module](mod.node, apphost.ModuleName)

	mod.auth.AddAuthorizer(&Authorizer{mod: mod})

	if mod.sdp != nil {
		mod.sdp.AddServiceDiscoverer(mod)
		mod.sdp.AddDataDiscoverer(mod)
	}

	mod.dir.AddDescriber(mod)

	mod.objects.AddObject(&user.NodeContract{})

	return nil
}

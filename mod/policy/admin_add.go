package policy

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

func (adm *Admin) add(term *admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("usage: policy add <policy> [policy_options]")
	}

	switch args[0] {
	case "always_linked":
		return adm.addAlwaysLinkedPolicy(term, args[1:])

	case "optimize_links":
		return adm.mod.AddPolicy(NewOptimizeLinksPolicy(adm.mod))

	case "reroute_conns":
		return adm.mod.AddPolicy(NewRerouteConnsPolicy(adm.mod))
	}

	return errors.New("unknown policy")
}

func (adm *Admin) addAlwaysLinkedPolicy(_ *admin.Terminal, args []string) error {
	var targets = make([]id.Identity, 0)
	for _, name := range args {
		target, err := adm.mod.node.Resolver().Resolve(name)
		if err != nil {
			return err
		}
		targets = append(targets, target)
	}

	var policy = NewAlwaysLinkedPolicy(adm.mod)
	if err := adm.mod.AddPolicy(policy); err != nil {
		return err
	}

	for _, target := range targets {
		if err := policy.AddIdentity(target); err != nil {
			return err
		}
	}

	return nil
}

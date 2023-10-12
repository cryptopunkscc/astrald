package policy

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/router"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Module struct {
	config    Config
	node      node.Node
	log       *log.Logger
	ctx       context.Context
	modRouter *router.Module
	policies  map[*RunningPolicy]struct{}
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, err := modules.Find[*admin.Module](mod.node.Modules()); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	mod.modRouter, _ = modules.Find[*router.Module](mod.node.Modules())

	if mod.config.AlwaysLinked != nil {
		if err := mod.addAlwaysLinkedPolicyFromConfig(mod.config.AlwaysLinked); err != nil {
			mod.log.Errorv(0, "error adding always_linked policy from config: %v", err)
		}
	}

	if mod.config.OptimizeLinks != nil {
		if err := mod.addOptimizeLinksPolicyFromConfig(mod.config.OptimizeLinks); err != nil {
			mod.log.Errorv(0, "error adding optimize_links policy from config: %v", err)
		}
	}

	if mod.config.RerouteConns != nil {
		mod.AddPolicy(NewRerouteConnsPolicy(mod))
	}

	<-ctx.Done()
	return nil
}

func (mod *Module) addAlwaysLinkedPolicyFromConfig(cfg *ConfigAlwaysLinked) error {
	policy := NewAlwaysLinkedPolicy(mod)
	if err := mod.AddPolicy(policy); err != nil {
		return err
	}

	for _, name := range cfg.Targets {
		target, err := mod.node.Resolver().Resolve(name)
		if err != nil {
			mod.log.Error("always_linked: error resolving %v: %v", name, err)
			continue
		}

		err = policy.AddIdentity(target)
		if err != nil {
			mod.log.Error("always_linked: error adding %v: %v", target, err)
		}
	}

	return nil
}

func (mod *Module) addOptimizeLinksPolicyFromConfig(cfg *ConfigOptimizeLinks) error {
	policy := NewOptimizeLinksPolicy(mod)

	return mod.AddPolicy(policy)
}

func (mod *Module) addRerouteConnsPolicyFromConfig(cfg *ConfigRerouteConns) error {
	policy := NewRerouteConnsPolicy(mod)

	return mod.AddPolicy(policy)
}

func (mod *Module) AddPolicy(policy Policy) error {
	running := RunPolicy(mod.ctx, policy)

	mod.policies[running] = struct{}{}

	return nil
}

func (mod *Module) AlwaysLinkedPolicy() *AlwaysLinkedPolicy {
	for p := range mod.policies {
		if p, ok := p.Policy.(*AlwaysLinkedPolicy); ok {
			return p
		}
	}
	return nil
}

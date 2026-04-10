package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ user.Module = &Module{}

type Module struct {
	Deps
	ctx    *astral.Context
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *DB
	ops    ops.Set

	activeContract *auth.SignedContract

	sibs sig.Map[string, Sibling]
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-mod.Scheduler.Ready()

	activeContractFollow := mod.config.ActiveContract.Follow(ctx)
	mod.setActiveContract(<-activeContractFollow)
	go func() {
		for contract := range activeContractFollow {
			mod.setActiveContract(contract)
		}
	}()

	mod.runSiblingLinker()
	<-ctx.Done()

	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return user.ModuleName
}

func (mod *Module) runSiblingLinker() {
	for _, node := range mod.LocalSwarm() {
		if node.IsEqual(mod.node.Identity()) {
			continue
		}

		_, ok := mod.sibs.Get(node.String())
		if ok {
			continue
		}

		maintainLinkAction := mod.NewMaintainLinkTask(node)
		scheduledAction, err := mod.Scheduler.Schedule(maintainLinkAction)
		if err != nil {
			mod.log.Error("error scheduling maintain link action: %v for node %v", err, node)
			continue
		}

		mod.addSibling(node, scheduledAction.Cancel)
	}
}

func (mod *Module) GetSwarmJoinRequestPolicy() user.SwarmJoinRequestPolicy {
	return mod.SwarmJoinRequestAcceptAll
}

var _ user.SwarmJoinRequestPolicy = (*Module)(nil).SwarmJoinRequestAcceptAll

func (mod *Module) SwarmJoinRequestAcceptAll(requester *astral.Identity) bool {
	mod.log.Info("Accepting %v join request into swarm", requester)
	return true
}

func (mod *Module) GetSwarmInvitePolicy() user.SwarmInvitePolicy {
	return mod.SwarmInviteAcceptAll
}

var _ user.SwarmInvitePolicy = (*Module)(nil).SwarmInviteAcceptAll

func (mod *Module) SwarmInviteAcceptAll(invitee *astral.Identity, contract *auth.Contract) bool {
	mod.log.Info("Accepting invitation from %v for %v join swarm till %v", invitee, contract.Subject, contract.ExpiresAt)
	return true
}

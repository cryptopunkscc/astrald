package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

var _ core.QueryPreprocessor = &Module{}

func (mod *Module) PreprocessQuery(qm *core.QueryModifier) error {
	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity()).ExcludeZone(astral.ZoneNetwork)

	// if the query comes from a locally hosted app, attach a contract with the app
	contracts, _ := mod.Auth.SignedContracts().
		WithIssuer(qm.Query().Caller).
		WithSubject(mod.node.Identity()).
		WithAction(&nodes.RelayForAction{}).
		Find(ctx)
	if len(contracts) > 0 {
		qm.Attach(contracts[0])
	}

	// if the query targets an app, find its hosts
	contracts, _ = mod.Auth.SignedContracts().
		WithIssuer(qm.Query().Target).
		WithAction(&nodes.RelayForAction{}).
		Find(ctx)

	for _, contract := range contracts {
		if contract.Subject.IsEqual(mod.node.Identity()) {
			continue
		}
		qm.AddRelay(contract.Subject)
	}

	return nil
}

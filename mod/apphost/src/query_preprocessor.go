package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

var _ core.QueryPreprocessor = &Module{}

func (mod *Module) PreprocessQuery(qm *core.QueryModifier) error {
	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity()).ExcludeZone(astral.ZoneNetwork)

	// if the query comes from a locally hosted app, attach a contract with the app
	contracts, _ := mod.db.FindActiveAppContractsByAppAndHost(qm.Query().Caller, mod.node.Identity())
	if len(contracts) > 0 {
		// query is coming from a locally hosted app
		contract, err := objects.Load[*apphost.AppContract](ctx, mod.Objects.Root(), contracts[0].ObjectID, nil)
		if err != nil {
			mod.log.Errorv(1, "cannot load app contract: %v", err)
		}

		if contract != nil {
			qm.Attach(contract)
		}
	}

	// if the query targets an app, find its hosts
	contracts, _ = mod.db.FindActiveAppContractsByApp(qm.Query().Target)

	for _, contract := range contracts {
		if contract.HostID.IsEqual(mod.node.Identity()) {
			continue
		}
		qm.AddRelay(contract.HostID)
	}

	return nil
}

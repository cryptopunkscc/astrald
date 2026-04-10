package auth

import (
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) StoreContract(ctx *astral.Context, sc *auth.SignedContract) error {
	objectID, err := objects.Save(ctx, sc, mod.Objects.WriteDefault())
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return mod.db.storeContract(
		objectID,
		sc.Issuer,
		sc.Subject,
		sc.ExpiresAt.Time(),
	)
}

func (mod *Module) FindContractsWithActor(ctx *astral.Context, actor *astral.Identity) ([]*auth.SignedContract, error) {
	rows, err := mod.db.findActiveContractsBySubject(actor)
	if err != nil {
		return nil, err
	}
	var result []*auth.SignedContract
	for _, row := range rows {
		sc, err := objects.Load[*auth.SignedContract](ctx, mod.Objects.ReadDefault(), row.ObjectID)
		if err != nil {
			continue
		}
		result = append(result, sc)
	}
	return result, nil
}

func (mod *Module) FindContractsWithIssuer(ctx *astral.Context, issuer *astral.Identity) ([]*auth.SignedContract, error) {
	rows, err := mod.db.findActiveContractsByIssuer(issuer)
	if err != nil {
		return nil, err
	}
	var result []*auth.SignedContract
	for _, row := range rows {
		sc, err := objects.Load[*auth.SignedContract](ctx, mod.Objects.ReadDefault(), row.ObjectID)
		if err != nil {
			continue
		}
		result = append(result, sc)
	}
	return result, nil
}

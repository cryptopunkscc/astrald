package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) SaveSignedRevocationContract(revocation *user.SignedNodeContractRevocation, contract *user.SignedNodeContract) (err error) {
	revocationID, err := astral.ResolveObjectID(revocation)
	if err != nil {
		return
	}

	if mod.db.ContractRevocationExists(revocationID) {

		return nil
	}

	nodeContract, err := mod.GetNodeContract(revocation.ContractID)
	if err != nil {
		return err
	}

	err = mod.ValidateNodeContractRevocation(revocation, contract)
	if err != nil {
		return err
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	// NOTE: ask about error handling in objects.Save
	_, err = objects.Save(ctx, revocation, mod.Objects.Root())
	if err != nil {
		return err
	}

	err = mod.db.Create(&dbNodeContractRevocation{
		ObjectID:   revocationID,
		ContractID: revocation.ContractID,
		ExpiresAt:  revocation.ExpiresAt.Time().UTC(),
		CreatedAt:  revocation.CreatedAt.Time().UTC(),
	}).Error
	if err != nil {
		return err
	}

	// Remove any sibling links for the node as its contract is revoked (no point of maintaining connection)
	mod.removeSibling(nodeContract.NodeID)

	return nil
}

func (mod *Module) ValidateNodeContractRevocation(revocation *user.SignedNodeContractRevocation, contract *user.SignedNodeContract) error {
	if revocation.Revoker.ID.IsZero() {
		return user.ErrNodeContractRevocationInvalid
	}

	if len(revocation.Revoker.Sig) == 0 {
		return user.ErrNodeContractRevocationInvalid
	}

	ok := mod.Auth.Authorize(revocation.Revoker.ID, user.ActionRevokeContract, contract)
	if !ok {
		return user.ErrNodeCannotRevokeContract
	}
	// verify user signature
	err := mod.Keys.VerifyASN1(revocation.Revoker.ID, revocation.Hash(), revocation.Revoker.Sig)
	if err != nil {
		return user.ErrContractInvalidSignature
	}

	return nil
}

func (mod *Module) LoadNodeContractRevocation(revocationID *astral.ObjectID) (*user.SignedNodeContractRevocation, error) {
	// fast fail so we dont need to load the contract if it does not exist in db
	if !mod.db.ContractExists(revocationID) {
		return nil, user.ErrContractNotExists
	}

	return objects.Load[*user.SignedNodeContractRevocation](
		mod.ctx,
		mod.Objects.Root(),
		revocationID,
		mod.Objects.Blueprints(),
	)
}

func (mod *Module) FindNodeContractRevocation(revocationID *astral.ObjectID) (r *user.NodeContractRevocation, err error) {
	dbRecord, err := mod.db.FindNodeContractRevocation(revocationID)
	if err != nil {
		return nil, err
	}

	return &user.NodeContractRevocation{
		ContractID: dbRecord.ContractID,
		CreatedAt:  astral.Time(dbRecord.CreatedAt),
		ExpiresAt:  astral.Time(dbRecord.ExpiresAt),
	}, nil
}

package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) SaveSignedRevocationContract(c *user.SignedNodeContractRevocation) (err error) {
	revocationID, err := astral.ResolveObjectID(c)
	if err != nil {
		return
	}

	if mod.db.ContractRevocationExists(revocationID) {
		return nil
	}

	nodeContract, err := mod.GetNodeContract(c.ContractID)
	if err != nil {
		return err
	}

	err = mod.ValidateNodeContractRevocation(c)
	if err != nil {
		return err
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	// NOTE: ask about error handling in objects.Save
	objects.Save(ctx, c, mod.Objects.Root())

	err = mod.db.Create(&dbNodeContractRevocation{
		ObjectID:   revocationID,
		UserID:     c.UserID,
		ContractID: c.ContractID,
		ExpiresAt:  c.ExpiresAt.Time().UTC(),
		StartsAt:   c.StartsAt.Time().UTC(),
	}).Error
	if err != nil {
		return err
	}

	// Remove any sibling links for the node as its contract is revoked (no point of maintaining connection)
	mod.removeSibling(nodeContract.NodeID)

	return nil
}

func (mod *Module) ValidateNodeContractRevocation(revocation *user.SignedNodeContractRevocation) error {
	if revocation.UserID.IsZero() {
		return user.ErrNodeContractRevocationInvalid
	}

	// verify user signature
	err := mod.Keys.VerifyASN1(revocation.UserID, revocation.Hash(), revocation.UserSig)
	if err != nil {
		// FIXME: invalid signature error (its not present on this branch currenlty)
		return user.ErrContractInvalidSignature
	}

	return nil
}

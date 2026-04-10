package auth

import (
	"fmt"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

func (mod *Module) VerifyContract(sc *auth.SignedContract) error {
	return mod.verifySignedContract(sc)
}

func (mod *Module) SignContract(ctx *astral.Context, c *auth.Contract) (*auth.SignedContract, error) {
	signed := &auth.SignedContract{Contract: c}

	var err error
	signed.IssuerSig, err = mod.signAs(ctx, secp256k1.FromIdentity(c.Issuer), c)
	if err != nil {
		return nil, fmt.Errorf("sign as issuer: %w", err)
	}

	signed.SubjecSig, err = mod.signAs(ctx, secp256k1.FromIdentity(c.Subject), c)
	if err != nil {
		return nil, fmt.Errorf("sign as subject: %w", err)
	}

	return signed, nil
}

// signAs signs the contract using the best available scheme for the given key.
// Prefers BIP137 (hardware wallet text signing) over ASN1.
func (mod *Module) signAs(ctx *astral.Context, key *crypto.PublicKey, c *auth.Contract) (*crypto.Signature, error) {
	schemes := mod.Crypto.AvailableSchemes(key)

	if slices.Contains(schemes, crypto.SchemeBIP137) {
		signer, err := mod.Crypto.TextObjectSigner(key)
		if err != nil {
			return nil, err
		}
		return signer.SignTextObject(ctx, c)
	}

	if slices.Contains(schemes, crypto.SchemeASN1) {
		signer, err := mod.Crypto.ObjectSigner(key)
		if err != nil {
			return nil, err
		}
		return signer.SignObject(ctx, c)
	}

	return nil, fmt.Errorf("no signing scheme available for key %v", key)
}

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

func (mod *Module) FindContracts(ctx *astral.Context, q auth.ContractQuery) ([]*auth.SignedContract, error) {
	var rows []*dbContract
	var err error

	switch {
	case q.IssuerID != nil:
		rows, err = mod.db.findActiveContractsByIssuer(q.IssuerID)
	case q.SubjectID != nil:
		rows, err = mod.db.findActiveContractsBySubject(q.SubjectID)
	default:
		return nil, fmt.Errorf("ContractQuery requires IssuerID or SubjectID")
	}
	if err != nil {
		return nil, err
	}

	var result []*auth.SignedContract
	for _, row := range rows {
		sc, err := objects.Load[*auth.SignedContract](ctx, mod.Objects.ReadDefault(), row.ObjectID)
		if err != nil {
			continue
		}

		if q.Action != "" {
			if !hasPermit(sc, q.Action) {
				continue
			}
		}

		result = append(result, sc)
	}

	return result, nil
}

func hasPermit(sc *auth.SignedContract, action string) bool {
	for _, p := range sc.Permits {
		if string(p.Action) == action {
			return true
		}
	}
	return false
}

func (mod *Module) Ban(ctx *astral.Context, identity *astral.Identity) error {
	return mod.db.addBan(identity)
}

func (mod *Module) Unban(ctx *astral.Context, identity *astral.Identity) error {
	return mod.db.removeBan(identity)
}

func (mod *Module) IsBanned(identity *astral.Identity) bool {
	return mod.db.isBanned(identity)
}

package apphost

import (
	"fmt"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

func (mod *Module) SignAppContract(ctx *astral.Context, c *apphost.AppContract) (*apphost.SignedAppContract, error) {
	signed := &apphost.SignedAppContract{AppContract: c}

	// sign as host (node) — ASN1 hash-based
	hostSigner, err := mod.Crypto.ObjectSigner(secp256k1.FromIdentity(c.HostID))
	if err != nil {
		return nil, fmt.Errorf("sign as host: %w", err)
	}
	signed.HostSig, err = hostSigner.SignObject(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("sign as host: %w", err)
	}

	// sign as app — ASN1 hash-based
	appSigner, err := mod.Crypto.ObjectSigner(secp256k1.FromIdentity(c.AppID))
	if err != nil {
		return nil, fmt.Errorf("sign as app: %w", err)
	}
	signed.AppSig, err = appSigner.SignObject(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("sign as app: %w", err)
	}

	return signed, nil
}

func (mod *Module) ActiveLocalAppContracts() (list []*apphost.SignedAppContract, err error) {
	contracts, err := mod.db.FindActiveAppContractsByHost(mod.node.Identity())
	if err != nil {
		return
	}

	for _, dbContract := range contracts {
		signed, err := objects.Load[*apphost.SignedAppContract](nil, mod.Objects.ReadDefault(), dbContract.ObjectID)
		if err != nil {
			mod.log.Errorv(2, "error loading contract %v: %v", dbContract.ObjectID, err)
			continue
		}
		list = append(list, signed)
	}

	return
}

func (mod *Module) verifySignedAppContract(signed *apphost.SignedAppContract) error {
	switch {
	case signed.IsNil():
		return fmt.Errorf("nil contract")
	case signed.HostSig == nil:
		return fmt.Errorf("host signature is missing")
	case signed.AppSig == nil:
		return fmt.Errorf("app signature is missing")
	case signed.SignableHash() == nil:
		return fmt.Errorf("cannot compute contract hash")
	}

	hash := signed.SignableHash()

	if err := mod.Crypto.VerifyHashSignature(
		secp256k1.FromIdentity(signed.HostID),
		signed.HostSig,
		hash,
	); err != nil {
		return fmt.Errorf("host sig: %w", err)
	}

	if err := mod.Crypto.VerifyHashSignature(
		secp256k1.FromIdentity(signed.AppID),
		signed.AppSig,
		hash,
	); err != nil {
		return fmt.Errorf("app sig: %w", err)
	}

	return nil
}

// isActive returns true if the contract is active, i.e., the signatures are valid, and its conditions are met (such
// as start and expiry date).
func (mod *Module) isActive(signed *apphost.SignedAppContract) bool {
	switch {
	case signed.StartsAt.Time().After(time.Now()):
		return false // hasn't started yet
	case signed.ExpiresAt.Time().Before(time.Now()):
		return false // has expired
	case mod.verifySignedAppContract(signed) != nil:
		return false // invalid signatures
	}

	return true
}

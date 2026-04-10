package auth

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

func (mod *Module) verifySignedContract(sc *auth.SignedContract) error {
	switch {
	case sc.IssuerSig == nil:
		return errors.New("issuer signature is missing")
	case sc.SubjecSig == nil:
		return errors.New("subject signature is missing")
	}

	if err := mod.verifySig(secp256k1.FromIdentity(sc.Issuer), sc.IssuerSig, sc.Contract); err != nil {
		return fmt.Errorf("issuer sig: %w", err)
	}

	if err := mod.verifySig(secp256k1.FromIdentity(sc.Subject), sc.SubjecSig, sc.Contract); err != nil {
		return fmt.Errorf("subject sig: %w", err)
	}

	return nil
}

func (mod *Module) verifySig(key *crypto.PublicKey, sig *crypto.Signature, contract *auth.Contract) error {
	switch sig.Scheme {
	case crypto.SchemeASN1:
		return mod.Crypto.VerifyObjectSignature(key, sig, contract)
	case crypto.SchemeBIP137:
		return mod.Crypto.VerityTextObjectSignature(key, sig, contract)
	default:
		return fmt.Errorf("unsupported signature scheme: %s", sig.Scheme)
	}
}

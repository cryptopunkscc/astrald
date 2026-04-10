package auth

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

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

// signAs signs the contract. Tries ASN1 first, falls back to BIP137.
func (mod *Module) signAs(ctx *astral.Context, key *crypto.PublicKey, c *auth.Contract) (*crypto.Signature, error) {
	if signer, err := mod.Crypto.ObjectSigner(key); err == nil {
		return signer.SignObject(ctx, c)
	}

	if signer, err := mod.Crypto.TextObjectSigner(key); err == nil {
		return signer.SignTextObject(ctx, c)
	}

	return nil, fmt.Errorf("no signing scheme available for key %v", key)
}

func (mod *Module) VerifyContract(sc *auth.SignedContract) error {
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

package auth

import (
	"errors"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

func (mod *Module) SignIssuer(ctx *astral.Context, signed *auth.SignedContract) (*crypto.Signature, error) {
	if signed.IssuerSig != nil {
		return signed.IssuerSig, nil
	}

	sig, err := mod.signAs(ctx, secp256k1.FromIdentity(signed.Issuer), signed.Contract)
	if err != nil {
		return nil, fmt.Errorf("sign as issuer: %w", err)
	}

	signed.IssuerSig = sig
	return sig, nil
}

func (mod *Module) SignSubject(ctx *astral.Context, signed *auth.SignedContract) (*crypto.Signature, error) {
	if signed.SubjecSig != nil {
		return signed.SubjecSig, nil
	}

	sig, err := mod.signAs(ctx, secp256k1.FromIdentity(signed.Subject), signed.Contract)
	if err != nil {
		return nil, fmt.Errorf("sign as subject: %w", err)
	}

	signed.SubjecSig = sig
	return sig, nil
}

func (mod *Module) SignContract(ctx *astral.Context, signed *auth.SignedContract) (*auth.SignedContract, error) {
	_, err := mod.SignIssuer(ctx, signed)
	if err != nil {
		return nil, err
	}

	_, err = mod.SignSubject(ctx, signed)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

func (mod *Module) signAs(ctx *astral.Context, key *crypto.PublicKey, c *auth.Contract) (*crypto.Signature, error) {
	// todo: We could check what schemas are available for the given key.
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

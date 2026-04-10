package auth

import (
	"errors"

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
	case sc.IssuerSig.Scheme != crypto.SchemeASN1:
		return errors.New("issuer signature scheme is not supported")
	case sc.SubjecSig.Scheme != crypto.SchemeASN1:
		return errors.New("subject signature scheme is not supported")
	}

	if err := mod.Crypto.VerifyObjectSignature(
		secp256k1.FromIdentity(sc.Issuer),
		sc.IssuerSig,
		sc.Contract,
	); err != nil {
		return err
	}

	return mod.Crypto.VerifyObjectSignature(
		secp256k1.FromIdentity(sc.Subject),
		sc.SubjecSig,
		sc.Contract,
	)
}

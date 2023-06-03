package proto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
)

const (
	CmdShift = "shift"
	CmdQuery = "query"
)

type Cmd struct {
	Cmd string `cslq:"[c]c"`
}

type QueryParams struct {
	Query string `cslq:"[c]c"`
}

type ShiftParams struct {
	Identity  id.Identity `cslq:"v"`
	Signature []byte      `cslq:"[c]c"`
}

func (shift ShiftParams) Verify(delegate id.Identity) bool {
	var hash = sha256.New()
	cslq.Encode(hash, "v", shift.Identity)
	cslq.Encode(hash, "v", delegate)
	sum := hash.Sum(nil)

	return ecdsa.VerifyASN1(shift.Identity.PublicKey().ToECDSA(), sum, shift.Signature)
}

func BuildShiftParams(identity id.Identity, delegate id.Identity) (ShiftParams, error) {
	if identity.PrivateKey() == nil {
		return ShiftParams{}, errors.New("private key missing")
	}

	var hash = sha256.New()
	cslq.Encode(hash, "v", identity)
	cslq.Encode(hash, "v", delegate)
	sum := hash.Sum(nil)

	sig, err := ecdsa.SignASN1(rand.Reader, identity.PrivateKey().ToECDSA(), sum)
	if err != nil {
		return ShiftParams{}, err
	}

	return ShiftParams{
		Identity:  identity,
		Signature: sig,
	}, nil
}

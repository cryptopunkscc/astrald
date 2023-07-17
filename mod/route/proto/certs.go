package proto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

const RelayCertPrefix = "astral.security.relay_certificate"
const RelayCertDefaultValidity = time.Minute

type RelayCert struct {
	Identity  id.Identity `cslq:"v"`
	Relay     id.Identity `cslq:"v"`
	ExpiresAt cslq.Time   `cslq:"v"`
	Signature []byte      `cslq:"[c]c"`
}

func NewRelayCert(identity id.Identity, relay id.Identity) *RelayCert {
	var cert = &RelayCert{
		Identity:  identity,
		Relay:     relay,
		ExpiresAt: cslq.Time(time.Now().Add(RelayCertDefaultValidity)),
	}
	if cert.Identity.PrivateKey() != nil {
		cert.Sign()
	}
	return cert
}

func (cert *RelayCert) Sign() (err error) {
	if cert.Identity.PrivateKey() == nil {
		return errors.New("private key missing")
	}

	if cert.Relay.IsZero() {
		return errors.New("relay missing")
	}

	cert.Signature, err = ecdsa.SignASN1(rand.Reader, cert.Identity.PrivateKey().ToECDSA(), cert.sum())

	return
}

func (cert *RelayCert) Verify() error {
	if time.Now().After(cert.ExpiresAt.Time()) {
		return errors.New("cert expired")
	}
	if !cert.verifySignature() {
		return errors.New("invalid signature")
	}

	return nil
}

func (cert *RelayCert) verifySignature() bool {
	if cert.Signature == nil {
		return false
	}

	return ecdsa.VerifyASN1(cert.Identity.PublicKey().ToECDSA(), cert.sum(), cert.Signature)
}

func (cert *RelayCert) sum() []byte {
	var hash = sha256.New()
	var enc = cslq.NewEncoder(hash)
	enc.Encodef("[c]cvvv", RelayCertPrefix, cert.Identity, cert.Relay, cert.ExpiresAt)
	return hash.Sum(nil)
}

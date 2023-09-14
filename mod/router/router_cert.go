package router

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"time"
)

const RouterCertType = "core.routing.router_certificate"

// RouterCert is a certificate stating that the Router can forward unencrypted traffic of the Issuer.
type RouterCert struct {
	IssuerIdentity id.Identity
	RouterIdentity id.Identity
	ExpiresAt      time.Time
	Signature      []byte
}

// NewRouterCert returns a new router certificate. If IssuerIdentity is a private key, certificate will be
// automatically signed.
func NewRouterCert(issuerIdentity id.Identity, routerIdentity id.Identity, expiresAt time.Time) *RouterCert {
	var cert = &RouterCert{
		IssuerIdentity: issuerIdentity,
		RouterIdentity: routerIdentity,
		ExpiresAt:      expiresAt,
	}

	// sign if possible
	_ = cert.Sign()

	return cert
}

// Sign signs the certificate using IssuerIdentity
func (cert *RouterCert) Sign() (err error) {
	if cert.IssuerIdentity.PrivateKey() == nil {
		return errors.New("issuer private key missing")
	}

	if cert.RouterIdentity.IsZero() {
		return errors.New("invalid router identity")
	}

	cert.Signature, err = ecdsa.SignASN1(rand.Reader, cert.IssuerIdentity.PrivateKey().ToECDSA(), cert.sha256())

	return err
}

// Verify verifies the signature of the certificate
func (cert *RouterCert) Verify() bool {
	if cert.Signature == nil {
		return false
	}

	return ecdsa.VerifyASN1(cert.IssuerIdentity.PublicKey().ToECDSA(), cert.sha256(), cert.Signature)
}

// Validate checks the validity of the certificate by checking its expiration time and signature
func (cert *RouterCert) Validate() error {
	if time.Now().After(cert.ExpiresAt) {
		return errors.New("certificate expired")
	}

	if !cert.Verify() {
		return errors.New("invalid signature")
	}

	return nil
}

func (cert *RouterCert) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("<x41x44x43x30><[c]c>vvv[c]c",
		RouterCertType,
		cert.IssuerIdentity,
		cert.RouterIdentity,
		cslq.Time(cert.ExpiresAt),
		cert.Signature,
	)
}

func (cert *RouterCert) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var expiresAt cslq.Time
	err := dec.Decodef("<x41x44x43x30><[c]c>vvv[c]c",
		RouterCertType,
		&cert.IssuerIdentity,
		&cert.RouterIdentity,
		&expiresAt,
		&cert.Signature,
	)
	cert.ExpiresAt = expiresAt.Time()
	return err
}

func (cert *RouterCert) sha256() []byte {
	var hash = sha256.New()
	var enc = cslq.NewEncoder(hash)
	err := enc.Encodef(
		"<[c]c>vvv",
		RouterCertType,
		cert.IssuerIdentity,
		cert.RouterIdentity,
		cslq.Time(cert.ExpiresAt),
	)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}

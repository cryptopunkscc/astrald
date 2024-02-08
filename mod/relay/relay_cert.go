package relay

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/lib/adc"
	"time"
)

type Direction string

const (
	Inbound             = Direction("inbound")
	Outbound            = Direction("outbound")
	Both                = Direction("both")
	DefaultCertDuration = 100 * 365 * 24 * time.Hour
)

type RelayCert struct {
	TargetID  id.Identity
	RelayID   id.Identity
	Direction Direction
	ExpiresAt time.Time
	TargetSig []byte
	RelaySig  []byte
}

func (cert *RelayCert) Hash() []byte {
	var hash = sha256.New()
	var err = cslq.Encode(hash,
		"[c]cvv[c]cv",
		RelayCertType,
		cert.TargetID,
		cert.RelayID,
		cert.Direction,
		cslq.Time(cert.ExpiresAt),
	)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}

// Validate checks if the certificate is valid, i.e. it hasn't expired and signatures are valid
func (cert *RelayCert) Validate() error {
	switch {
	case cert.ExpiresAt.Before(time.Now()):
		return errors.New("certificate expired")
	case cert.TargetID.IsEqual(cert.RelayID):
		return errors.New("relay and target cannot be equal")
	}

	return cert.Verify()
}

// Verify verifies signatures of the certificate
func (cert *RelayCert) Verify() error {
	return errors.Join(cert.VerifyRelay(), cert.VerifyTarget())
}

// VerifyRelay verfies relay signature
func (cert *RelayCert) VerifyRelay() error {
	switch {
	case cert.RelaySig == nil:
		return errors.New("relay signature missing")
	case cert.RelayID.IsZero():
		return errors.New("relay identity missing")
	case !ecdsa.VerifyASN1(
		cert.RelayID.PublicKey().ToECDSA(),
		cert.Hash(),
		cert.RelaySig,
	):
		return errors.New("relay signature invalid")
	}

	return nil
}

// VerifyTarget verifies target signature
func (cert *RelayCert) VerifyTarget() error {
	switch {
	case cert.TargetSig == nil:
		return errors.New("target signature missing")
	case cert.TargetID.IsZero():
		return errors.New("target identity missing")
	case !ecdsa.VerifyASN1(
		cert.TargetID.PublicKey().ToECDSA(),
		cert.Hash(),
		cert.TargetSig,
	):
		return errors.New("target signature invalid")
	}

	return nil
}

func (cert RelayCert) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encodef("vv[c]cv[c]c[c]c",
		cert.TargetID,
		cert.RelayID,
		cert.Direction,
		cslq.Time(cert.ExpiresAt),
		cert.TargetSig,
		cert.RelaySig,
	)
}

func (cert *RelayCert) UnmarshalCSLQ(dec *cslq.Decoder) error {
	var expiresAt cslq.Time
	err := dec.Decodef("vv[c]cv[c]c[c]c",
		&cert.TargetID,
		&cert.RelayID,
		&cert.Direction,
		&expiresAt,
		&cert.TargetSig,
		&cert.RelaySig,
	)
	cert.ExpiresAt = expiresAt.Time()
	return err
}

func UnmarshalCert(p []byte) (*RelayCert, error) {
	var r = bytes.NewReader(p)

	var t adc.Header
	var cert RelayCert

	var err = cslq.Decode(r, "vv", &t, &cert)
	if err != nil {
		return nil, err
	}

	if t != RelayCertType {
		return nil, errors.New("invalid data type")
	}

	return &cert, nil
}

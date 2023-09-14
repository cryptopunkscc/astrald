package router

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
)

// IdentityMachine renders the final identity by applying certificates to the initial identity
type IdentityMachine struct {
	identity id.Identity
}

// NewIdentityMachine returns a new instance of an IdentityMachine with the provided identity as its initial state
func NewIdentityMachine(identity id.Identity) *IdentityMachine {
	return &IdentityMachine{identity: identity}
}

// Apply applies a certificate to the current identity
func (m *IdentityMachine) Apply(certBytes []byte) error {
	// unmarshal certificate
	var cert RouterCert
	if err := cslq.Unmarshal(certBytes, &cert); err != nil {
		return err
	}

	// check if the certificate can be applied to current identity
	if !cert.RouterIdentity.IsEqual(m.identity) {
		return errors.New("certificate identity mismatch")
	}

	// validate the certificate
	if err := cert.Validate(); err != nil {
		return errors.New("caller provided an invalid certificate")
	}

	m.identity = cert.IssuerIdentity

	return nil
}

// Identity returns the current identity
func (m *IdentityMachine) Identity() id.Identity {
	return m.identity
}

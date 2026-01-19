package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	SchemeASN1   = "asn1"
	SchemeBIP137 = "bip137"
)

// Engine adds support for various cryptographic operations to the crypto module. Engines can
// implement any subset of Engine operations. Every operation has to verify that it supports
// the key type and scheme and return ErrUnsupported if it doesn't.
type Engine interface {
	// PublicKey derives the public key from the provided private key
	PublicKey(ctx *astral.Context, key *PrivateKey) (*PublicKey, error)

	// HashSigner returns a signer that will sign hashes using the provided key and scheme
	HashSigner(key *PublicKey, scheme string) (HashSigner, error)

	// VerifyHashSignature verifies a signature of a hash
	VerifyHashSignature(key *PublicKey, sig *Signature, hash []byte) error

	// MessageSigner returns a signer that will sign messages using the provided key and scheme
	MessageSigner(key *PublicKey, scheme string) (MessageSigner, error)

	// VerifyMessageSignature verifies a signature of a message
	VerifyMessageSignature(key *PublicKey, sig *Signature, msg string) error
}

type HashSigner interface {
	// SignHash generates a signature for the given hash
	SignHash(ctx *astral.Context, hash []byte) (*Signature, error)
}

type MessageSigner interface {
	// SignMessage generates a signature for the given message
	SignMessage(ctx *astral.Context, msg string) (*Signature, error)
}

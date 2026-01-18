package crypto

import "github.com/cryptopunkscc/astrald/astral"

const (
	SchemeASN1   = "asn1"
	SchemeBIP137 = "bip137"
)

// Engine is an interface for cryptographic operations
type Engine interface {
	// PublicKey derives the public key from the provided private key
	PublicKey(ctx *astral.Context, key *PrivateKey) (*PublicKey, error)

	// HashSigner returns a signer for the given signature scheme
	HashSigner(key *PublicKey, scheme string) (HashSigner, error)
	VerifyHashSignature(key *PublicKey, sig *Signature, hash []byte) error
	MessageSigner(key *PublicKey, scheme string) (MessageSigner, error)
	VerifyMessageSignature(key *PublicKey, sig *Signature, msg string) error
}

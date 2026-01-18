/*
Package crypto provides a module with cryptographic operations and objects.
*/
package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "crypto"
const DBPrefix = "crypto__"

type Module interface {
	// PrivateKeyID looks up the ObjectID of a private key corresponding to the given public key
	PrivateKeyID(*PublicKey) (*astral.ObjectID, error)

	// PrivateKey tries to load the private key corresponding to the given public key
	PrivateKey(ctx *astral.Context, key *PublicKey) (*PrivateKey, error)

	// PublicKey generates the corresponding public key for the given private key
	PublicKey(ctx *astral.Context, key *PrivateKey) (*PublicKey, error)

	// HashSigner returns a hash signer for the given public key and scheme
	HashSigner(key *PublicKey, scheme string) (HashSigner, error)

	// VerifyHashSignature verifies the signature of the given hash using the given public key
	VerifyHashSignature(key *PublicKey, sig *Signature, hash []byte) error

	MessageSigner(key *PublicKey, scheme string) (MessageSigner, error)
	VerifyMessageSignature(key *PublicKey, sig *Signature, msg string) error

	AddEngine(engine Engine)
}

type HashSigner interface {
	SignHash(ctx *astral.Context, hash []byte) (*Signature, error)
}

type MessageSigner interface {
	SignMessage(ctx *astral.Context, msg string) (*Signature, error)
}

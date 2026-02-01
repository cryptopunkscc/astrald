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

	// TextSigner returns a message signer for the given public key and scheme
	TextSigner(key *PublicKey, scheme string) (TextSigner, error)

	// VerifyTextSignature verifies a text signature
	VerifyTextSignature(signer *PublicKey, sig *Signature, text string) error

	// NodeSigner returns a hash signer for the local node
	NodeSigner() HashSigner

	// AddEngine adds a cryptographic engine to the module
	AddEngine(engine Engine)

	// ObjectSigner signs the hash of the given contract with ASN1
	ObjectSigner(*PublicKey) (ObjectSigner, error)

	// TextObjectSigner signs the text of the given contract with BIP-137
	TextObjectSigner(*PublicKey) (TextObjectSigner, error)

	VerifyObjectSignature(*PublicKey, *Signature, SignableObject) error

	VerityTextObjectSignature(*PublicKey, *Signature, SignableTextObject) error
}

type ObjectSigner interface {
	SignObject(*astral.Context, SignableObject) (*Signature, error)
}

type TextObjectSigner interface {
	SignTextObject(*astral.Context, SignableTextObject) (*Signature, error)
}

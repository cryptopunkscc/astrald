package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const (
	SchemeASN1   = "asn1"   // default for hash signatures
	SchemeBIP137 = "bip137" // default for text signatures
)

// KeyDeriver derives public keys from private keys.
type KeyDeriver interface {
	// PublicKey derives the public key from the provided private key.
	// Returns ErrUnsupported if the key type is not recognised.
	PublicKey(ctx *astral.Context, key *PrivateKey) (*PublicKey, error)
}

// HashVerifier verifies signatures of arbitrary hashes.
// Returns ErrInvalidSignature if the signature is cryptographically invalid,
// or nil if valid. The registry guarantees the (keyType, scheme) pair is valid,
// so ErrUnsupported should not be returned.
type HashVerifier interface {
	VerifyHash(key *PublicKey, sig *Signature, hash []byte) error
}

// TextVerifier verifies signatures of text messages.
// Returns ErrInvalidSignature if the signature is cryptographically invalid,
// or nil if valid.
type TextVerifier interface {
	VerifyText(key *PublicKey, sig *Signature, text string) error
}

// HashSignerFactory produces signers for hash signing.
// Takes a PrivateKey directly — the caller resolves it from the public key.
type HashSignerFactory interface {
	NewHashSigner(ctx *astral.Context, key *PrivateKey, scheme string) (HashSigner, error)
}

// TextSignerFactory produces signers for text message signing.
type TextSignerFactory interface {
	NewTextSigner(ctx *astral.Context, key *PrivateKey, scheme string) (TextSigner, error)
}

// HashSigner signs hashes.
type HashSigner interface {
	// SignHash generates a signature for the given hash.
	SignHash(ctx *astral.Context, hash []byte) (*Signature, error)
}

// TextSigner signs text messages.
type TextSigner interface {
	// SignText generates a signature for the given text.
	SignText(ctx *astral.Context, text string) (*Signature, error)
}

// EngineProvider is implemented by modules that provide crypto capabilities.
// The crypto module calls RegisterCryptoCapabilities during startup to
// register all available key types, schemes, and handlers.
type EngineProvider interface {
	RegisterCryptoCapabilities(ctx *astral.Context, reg *Registry)
}

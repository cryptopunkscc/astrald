# mod/secp256k1

Provides a `crypto.Engine` for secp256k1 keys: derives public keys, produces ASN.1 hash signers, and verifies ASN.1 hash signatures. Also exposes package-level helpers that convert between `astral.Identity` and `crypto.PublicKey` and a `secp256k1.new` op that generates a fresh private key.

## Dependencies

| Module | Why |
| --- | --- |
| `crypto` | engine auto-registered through `CryptoEngine()`; engine calls `Crypto.PrivateKey` to load the signing key from the local index |
| `astral.Node` | injected via `Deps` but not used by the engine itself; available for future ops |
| `core/assets` | `LoadYAML` reads the (currently empty) module config; `Database()` backs an unused `DB` placeholder |

## Flows

- Engine registration: loader builds the `Module` -> `mod/crypto.LoadDependencies` discovers `EngineProvider`s and calls `CryptoEngine()` -> `Engine{mod}` joins the crypto engine set.
- Public key derivation: `Engine.DerivePublicKey` delegates to package `secp256k1.PublicKey` which calls `secp256k1.PrivKeyFromBytes(...).PubKey().SerializeCompressed()`.
- Hash signer: `Engine.NewHashSigner` rejects non-`secp256k1` key types and non-`asn1` schemes -> `Crypto.PrivateKey(key)` resolves the local private key -> returns `NewHashSignerASN1(privateKey)` which holds an `ecdsa.PrivateKey` for `crypto/ecdsa.SignASN1` over `crypto/rand`.
- Hash verification: `Engine.VerifyHashSignature` checks key type/scheme -> `secp256k1.ParsePubKey` -> `ecdsa.VerifyASN1` -> returns `crypto.ErrInvalidSignature` on parse or verify failure.
- Op `secp256k1.new`: accepts a channel, sends `secp256k1.New()` which generates a fresh `*crypto.PrivateKey` with `Type = "secp256k1"`.
- Identity bridge: `secp256k1.FromIdentity(*astral.Identity)` returns a compressed `*crypto.PublicKey`; `secp256k1.Identity(*crypto.PublicKey)` parses the bytes back into `*astral.Identity`.

## Source

- `mod/secp256k1/module.go` - module name, `KeyType` constant, and `New`, `PublicKey`, `FromIdentity`, `Identity` helpers.
- `mod/secp256k1/asn1.go`, `hash_signer_asn1.go` - `SignASN1`/`VerifyASN1` package helpers and the `HashSignerASN1` type returned to `mod/crypto`.
- `mod/secp256k1/src/loader.go`, `module.go`, `deps.go`, `config.go`, `db.go` - registration, lifecycle, and (currently empty) config/DB wiring.
- `mod/secp256k1/src/engine.go` - the `crypto.Engine` implementation: `DerivePublicKey`, `NewHashSigner`, `VerifyHashSignature`.
- `mod/secp256k1/src/op_new.go` - `secp256k1.new` handler.
- `mod/secp256k1/client/` - typed client wrapper for `secp256k1.new`.

## Invariants

- Engine accepts only `key.Type == "secp256k1"` and `scheme == "asn1"`; other combinations return `crypto.ErrUnsupportedKeyType` / `ErrUnsupportedScheme` so `mod/crypto` keeps fanning out.
- `VerifyHashSignature` returns `crypto.ErrInvalidSignature` (terminal in the verify fan-out) on both `ParsePubKey` failure and ECDSA mismatch.
- `KeyType = "secp256k1"` is the canonical string used across `mod/crypto`, `mod/bip137sig`, and `mod/coldcard`.
- Compressed public-key serialization is the standard form; `FromIdentity` and `PublicKey` both emit 33-byte compressed keys.

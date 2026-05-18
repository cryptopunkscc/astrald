# mod/crypto

Signs and verifies hashes and text through pluggable crypto engines, and indexes private keys so callers can resolve a public key to a local signer. Owns the engine registry, local node key loading, signer lookup, object/text signing adapters, and the `crypto__private_keys` index.

## Dependencies

| Module | Why |
| --- | --- |
| `objects` | stores the node key in the system repository, loads indexed private keys through `ReadDefault`, and scans configured repositories for private-key objects |
| `dir` (opt) | injected in `Deps`; current crypto code does not call it |
| `secp256k1` | derives the default signing public key from a query caller identity in signing ops |
| `core` | discovers loaded `EngineProvider` modules during `LoadDependencies` and registers their crypto engines |
| `core/assets` | `LoadYAML` reads crypto config, `Database()` backs `crypto__private_keys`, and `Res().Read("node_key")` loads the local node private key |

## Flows

- Engine discovery: `LoadDependencies` injects `Objects` and optional `Dir` -> scans loaded modules for `crypto.EngineProvider` -> adds each returned engine to the in-memory engine set.
- Node key setup: loader reads `node_key` from resources -> decodes a `PrivateKey`; dependencies stage stores it in `Objects.System()` -> `indexPrivateKey` records its public key mapping.
- Repository key indexing: `Run` starts one goroutine per configured repository -> `repo.Scan(ctx, true)` -> skip object IDs larger than 4096 bytes -> load private keys -> derive public key through engines -> insert `crypto__private_keys` row.
- Private key lookup: `PrivateKeyID` marshals the public key text -> finds the matching row by `public_key`; `PrivateKey` then loads `KeyID` through `Objects.ReadDefault()` and requires a `PrivateKey` object.
- Public key derivation: `PublicKey` clones the engine set -> asks each engine to derive the public key -> first successful engine wins, otherwise returns `ErrUnsupported`.
- Hash signing: `crypto.sign_hash` defaults scheme to `asn1` and signer key to `secp256k1.FromIdentity(q.Caller())` -> optional key/hash query args override channel input -> `HashSigner` dispatches to the first engine that can sign.
- Text signing: `crypto.sign_text` follows the text-signer path and defaults to the caller identity key and BIP137-style text signatures.
- Signature verification: verify ops and module methods reject nil or empty public keys, signatures, and payloads -> clone engines -> return on first nil error -> return `ErrInvalidSignature` immediately -> continue past non-matching engine errors -> return `ErrUnsupported`.
- Object signing: `ObjectSigner` wraps `HashSigner(key, asn1).SignHash(SignableHash())`; `TextObjectSigner` wraps `TextSigner(key, bip137).SignText(formatSignableText(object))`.
- Node signing: `NodeSigner` builds the local secp256k1 public key from `node.Identity()` -> requires an ASN1 hash signer and panics if no engine provides one.

## Source

- `mod/crypto/module.go`, `engine.go`, `nil_engine.go`, `errors.go` - public module interface, engine contracts, nil engine, and sentinels.
- `mod/crypto/private_key.go`, `public_key.go`, `signature.go`, `hash.go`, `signable_object.go` - crypto object types, text encodings, and signable object contracts.
- `mod/crypto/src/loader.go`, `module.go`, `deps.go`, `config.go` - registration, node key loading, dependency injection, engine fan-out, and indexing lifecycle.
- `mod/crypto/src/db.go`, `db_private_key.go` - private-key index schema and lookup helpers.
- `mod/crypto/src/object_signer.go`, `text_object_signer.go` - object and text-object signing adapters.
- `mod/crypto/src/op_public_key.go`, `op_sign_hash.go`, `op_sign_text.go`, `op_verify_hash_signature.go`, `op_verify_text_signature.go` - query operation handlers.
- `mod/crypto/client/` - typed client wrappers for crypto operations.

## Surface

| What | Why it matters |
| --- | --- |
| `crypto.Engine` and `EngineProvider` | extension point used by algorithm modules to provide public-key derivation, signing, and verification |
| `HashSigner`, `TextSigner`, `ObjectSigner`, `TextObjectSigner` | signer abstractions used by auth, ether, and other modules |
| `crypto.public_key`, `crypto.sign_hash`, `crypto.sign_text` | query methods for deriving public keys and producing signatures |
| `crypto.verify_hash_signature`, `crypto.verify_text_signature` | query methods and module paths for validating signatures |
| `crypto__private_keys` | maps public-key text to private-key object IDs for local signing |

## Invariants

- Engines self-filter by key type/scheme; non-`ErrInvalidSignature` errors mean "not me".
- `NilEngine` returns stdlib `errors.ErrUnsupported`; dispatch skips it.
- `formatSignableText` reads `SignableHash()[0:15]`; objects must yield at least 15 hash bytes.
- Private-key resolution limited to keys indexed from node key or `crypto.repos` (default `[local, system, mem0]`).
- `NodeSigner` panics if no engine supplies `asn1` for the local secp256k1 identity.
- Auto-index ceiling: hard-coded `maxObjectSize = 4096`.
- Encodings: `PrivateKey` text `type:base64(key)`; `PublicKey` text `type:hex(key)`; `Signature` text `scheme:base64(data)`; `Hash` binary `Bytes8`, text/JSON hex.

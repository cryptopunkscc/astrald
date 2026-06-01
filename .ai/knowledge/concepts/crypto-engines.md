# Crypto Engines

## Purpose

`mod/crypto` is the node-wide signing surface. Modules do not call crypto libraries directly.

Private keys do not have to live in software. A Coldcard hardware wallet keeps keys on a physical
device that the OS cannot read. If signing were hard-coded to a library call, hardware keys would not
work.

All signing goes through the `mod/crypto` engine fan-out. `mod/crypto` delegates to the engine that
claims the key, whether that engine is an in-process secp256k1 library or a USB device reached over a
protocol.

## EngineProvider

Modules that provide a signing backend implement `EngineProvider`:

* `CryptoEngine() Engine`

`Engine` is an opaque `any`. The engine value implements any subset of the capability interfaces in
`mod/crypto/engine.go`:

* `PublicKeyDeriver` - derive a public key from a private key
* `HashSignerProvider` - return a `HashSigner` bound to a key and scheme
* `HashVerifier` - verify a hash signature
* `TextSignerProvider` - return a `TextSigner` bound to a key and scheme
* `TextVerifier` - verify a text signature

At startup, `mod/crypto.LoadDependencies` walks loaded modules, calls `CryptoEngine()` on each
`EngineProvider`, and adds the result to its engine set. Every public method dispatches per capability
via `dispatchResult` / `dispatchVerify` (see `mod/crypto/src/dispatch.go`):

* On sign / derive paths, the first engine that returns a non-error result wins. Any returned error is
  treated as "skip me, keep looking". If no engine matches, `mod/crypto` returns `ErrUnsupported`.
* On verify paths, any engine returning `nil` succeeds. `ErrInvalidSignature` is terminal — "the key is
  mine and the signature is wrong". Any other error means "skip me, keep looking". If no engine
  matches, `mod/crypto` returns `ErrUnsupported`.

Engines that do not implement the requested capability interface are silently skipped: capability is
detected by type assertion, not by error.

## Signing Paths

### Hash Signing

`HashSigner`, scheme `asn1`, signs an opaque byte hash. It is fast, efficient, and unreadable.

Use it for internal paths where only machines verify the result:

* node authentication
* object signatures
* any operation where human review is not required

### Text Signing

`TextSigner`, scheme `bip137`, signs a human-readable string. Before signing, `mod/crypto` formats
the content as:

```text
[<hash-prefix>] <human text>
```

A hardware wallet like Coldcard displays this string before the user approves. This is the consent
path: the user sees the commitment before the key is used.

Use it when a human must read and approve the commitment.

## When to use each path

* Default internal operations, including node auth and object signing: `HashSigner` /
  `ObjectSigner`.
* User-facing authorisations where a human should review the payload: `TextSigner` /
  `TextObjectSigner`.

Never call a crypto library directly. Always go through `mod/crypto` so the correct engine is
selected and hardware keys remain possible.

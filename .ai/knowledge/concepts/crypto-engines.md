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

At startup, providers register with `mod/crypto`. The crypto module stores the engines in a set and
tries them in order for each operation.

* `ErrUnsupported` means "not mine, keep looking". It is a routing signal, not a failure.
* `ErrInvalidSignature` means "this is mine and it is wrong". The fan-out stops immediately.

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

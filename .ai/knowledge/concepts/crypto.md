# Crypto Engines

`mod/crypto` is the node-wide signing boundary.

* No module calls a crypto library directly.
* All signing goes through the pluggable engine fan-out.
* Private keys may live outside software. For example, a Coldcard hardware
  wallet holds keys on a USB device the OS cannot read.
* `mod/crypto` delegates signing to the engine that claims the key. That engine
  may be an in-process library or a physical device.

## Engine Providers

Any module can provide a crypto engine:

* Implement `CryptoEngine() Engine`. `Engine` is an opaque `any`; the engine
  value implements any subset of the capability interfaces:
  `PublicKeyDeriver`, `HashSignerProvider`, `HashVerifier`,
  `TextSignerProvider`, `TextVerifier`.
* `mod/crypto` discovers providers during `LoadDependencies` and dispatches per
  capability with a type assertion. Engines lacking the requested capability
  are silently skipped.
* Sign / derive path: return any non-nil error to mean "not me, keep looking".
  The first engine that returns a result wins.
* Verify path: return `ErrInvalidSignature` to mean "the key is mine and the
  signature is wrong" — the fan-out stops immediately. Any other error means
  "not me, keep looking".

## Signing Paths

| Path | Scheme | When to use |
|---|---|---|
| Hash signing | `asn1` | Internal: node auth and object signatures; opaque and machine-verified |
| Text signing | `bip137` | User-facing: formats a human-readable string a hardware wallet displays for approval |

Text signing is the consent path. Use it when a human must read and approve what
they sign.

Invariant: call `mod/crypto` for signing. Do not call crypto libraries directly.

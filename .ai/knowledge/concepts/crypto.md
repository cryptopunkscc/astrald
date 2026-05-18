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

* Implement `CryptoEngine() Engine`.
* Register at startup.
* Return `ErrUnsupported` when the key is not owned by the engine. The fan-out
  continues.
* Return `ErrInvalidSignature` when the key is owned by the engine but the
  signature is wrong. The fan-out stops immediately.

## Signing Paths

| Path | Scheme | When to use |
|---|---|---|
| Hash signing | `asn1` | Internal: node auth and object signatures; opaque and machine-verified |
| Text signing | `bip137` | User-facing: formats a human-readable string a hardware wallet displays for approval |

Text signing is the consent path. Use it when a human must read and approve what
they sign.

Invariant: call `mod/crypto` for signing. Do not call crypto libraries directly.

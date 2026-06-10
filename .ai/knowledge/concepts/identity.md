# Identity

## Representation

`astral.Identity` is a compressed secp256k1 public key — see [common-types/identity](../../system/common-types/identity.md).

Invariants:

* Binary form: 33 bytes (all zero when zero).
* String form: 66 hex characters; the string `"anyone"` is also accepted on
  parse and is emitted by JSON marshaling of the zero value.
* `Anyone`: zero value; anonymous or wildcard identity.
* `IsZero()` is true for nil receivers and for empty public keys.
* `IsEqual` treats two zero identities as equal.

## Addressing

Identity is the address. There is no hostname.

Discovery uses directory resolution and endpoint advertisement.

## User Identity

A user identity represents a human operator.

It owns a **Swarm** of Node identities that share trust, routes, and assets
(`mod/user`).

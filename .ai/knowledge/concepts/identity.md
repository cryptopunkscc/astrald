# Identity

## Representation

`astral.Identity` is a compressed secp256k1 public key.

Invariants:

* Binary form: 33 bytes.
* String form: 66 hex characters.
* `Anyone`: zero value; anonymous or wildcard identity.

## Addressing

Identity is the address. There is no hostname.

Discovery uses directory resolution and endpoint advertisement.

## User Identity

A user identity represents a human operator.

It owns a **Swarm** of Node identities that share trust, routes, and assets
(`mod/user`).

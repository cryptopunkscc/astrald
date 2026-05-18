# Brontide

Authenticated encrypted transport between a raw `exonet.Conn` and the mux
layer.

## Purpose

* Brontide turns an untrusted byte stream into an authenticated encrypted
  connection.
* Both peers get cryptographic proof of the peer identity before the connection
  reaches the mux layer.

## Noise XK

* Astrald `Identities` are secp256k1 keys, the same curve Bitcoin uses.
* Noise XK natively supports secp256k1.
* The handshake key and Astral `Identity` key are the same key pair.
* There is no separate TLS certificate or split identity.
* Proving ownership of an `Identity` on the wire is the same proof used in other
  Astral contexts.

## Responder Identity

* In XK, the initiator knows the responder static key before the handshake
  starts.
* In Astrald, the responder static key is the node `Identity`.
* Dialing a node commits the initiator to the expected node `Identity`.
* The handshake enforces that identity and rejects impersonation.

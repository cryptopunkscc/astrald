astrald
=======

## Overview

**The Astral Network** is an abstract, general-purpose peer-to-peer network layer that runs on top of multiple transports. **astrald** is a daemon that provides easy access to the network via local interface. It provides connectivity, identity, authentication, encryption and a set of core protocols apps and agents can use to easily interact over the network. The full documentation is [here](https://github.com/cryptopunkscc/astral-docs/blob/master/README.md).

## Network architecture

All actors (nodes, apps, users and agents) on the network have a self-assigned identity based on a ECDSA key on the secp256k1 curve, the same as the Lightning Network or Bitcoin. Nodes form the core infrastructure of the network by establish encrypted and authenticated links with each other, exchanging certificates, gossip data and organizing complex p2p processes. Apps, agents and users access the network via the API exposed by the daemon via varoius transports - astral apphost protocol, HTTP, WSS, etc.

Actors send each other queries (similar to HTTP query path) over the network that can be rejected or accepted to establish a connection. The connection is then used to exchange data in form of Objects (or just a raw bytestream).

## Quick start

Install the daemon and the tooling:

```shellsession
go install \
  github.com/cryptopunkscc/astrald/cmd/astrald@latest \
  github.com/cryptopunkscc/astrald/cmd/astral-query@latest
```

Run the daemon:

```shellsession
$ astrald
```

The node will start and generate a new identity and an alias for itself. Once the node has started verify it is accepting local API calls:

```shellsession
$ astral-query localnode:.spec
```

## Generating user identity

### New identity

TODO

### Existing identity

TODO

### Hardware identity

Astral can use BIP-0137-compatible hardware wallets as the source for user identities. See the [coldcard module](mod/coldcard/README.md).
# Overview

## Mission

Astral is an abstract network that provides authenticated and encrypted
connections over a variety of physical networks. It provides simple and secure
connectivity interface, which automatically adapts to existing network
conditions. Its mission is to dramatically reduce the time it takes to build
robust peer-to-peer networks.

## Core Design

### Identities

The central concept in the Astral network is an Identity. An Identity is any
entity that can perform cryptographic actions, such as signing. An Identity
can represent a person, a piece of software or even an authority over some
data structure.

Technically, an Identity is implemented by a key pair. The public key acts
as Identity's identifier on the network, similar to an IP address. Currently
Astral uses ECDSA with secp256k1 curve.

### Services and queries

Some Identities can provide Services to other Identities. Services are
identified by Service names, which are represented by strings. For example,
an Identity might provide a service called `http`, which offers a website
delivered over the HTTP protocol.

To access a Service, the requesting Identity sends a Query to the responding
Identity over any means agreed upon by both parties (see Nodes below). If
the query is accepted, a bidirectional bytestream is established and used for
the purpose of the service. Such bytestream is called a Session.

An Identity and a Service name is somewhat similar to an IP address and a port
number.

### Nodes

Some identities are capable of establishing Links with other Identities.
Such Identity is called a Node.

A Link is an encrypted and authenticated bidirectional bytestream between
two Nodes. Nodes can use Links to send and execute Queries. Nodes can
agree on any protocol/scheme for managing Links, as long as it provides
encryption and authentication.

The standard Link implementation uses Noise XK protocol and can carry multiple
Sessions simultaneously using a multiplexer.

### Apps

Managing Links is a complex process that can be very difficult and
time-consuming. An Identity may choose to delegate routing and network
management to a Node. A Node that is capable of providing routing service
is called a Host and an Identity that uses these services is called an App.

In the reference implementation, the module `apphost` provides an interface
for Apps to route traffic through the Astral network.

### Users

A User is an Identity that cannot directly establish Links or use Services
for obvious reasons. Such Identity will only sign various documents that
can be used by other Identities.

In practice, a User will use their Identity to sign certificates for the Nodes
that they control, so that these Nodes can represent the User on the network.
For example, if a Node `S` is set up to provide a service only to a User `A`,
any Node signed by the User `A` will also be granted that access.

The User can grant time-limited certifactes to Nodes using a hardware signing
key to prevent their personal key from being stolen from a hacked device.
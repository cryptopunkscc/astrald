# Streams and Sessions in Astral

## Overview

Astral’s `mod/nodes` module provides the peer-to-peer networking layer through two primary abstractions: **streams** and **sessions**.

* **Streams** are authenticated transport channels between nodes.
* **Sessions** are logical, bidirectional data flows multiplexed over those streams.

Together, they form a robust foundation that allows nodes to communicate across changing network paths while keeping higher-level connections intact.

### Core Principles

* Multiple sessions share a single encrypted stream.
* Sessions are independent logical connections with isolated buffers.
* Migration enables moving sessions between streams transparently.
* Flow control and signaling are handled through a lightweight frame protocol.

---

## Architecture Overview

Astral’s connection stack can be viewed in layers:

```
Application / Modules
        ↓
Sessions (logical connections)
        ↓
Streams (authenticated multiplexed transport)
        ↓
Noise-secured link (TCP/KCP/Tor/etc.)
```

* **Streams** handle encryption, framing, and peer liveness.
* **Sessions** handle application data flow and lifecycle management.

This separation ensures that transport changes (e.g., switching networks) do not affect the logical session layer.

---

## Stream Lifecycle

Streams are authenticated links between two node identities.
Each stream is established through a Noise handshake and registered in the peer set.

Streams:

* Carry frames for all active sessions with a given peer.
* Maintain liveness via ping/pong exchanges.
* Terminate all sessions upon disconnection.

A single identity may have multiple concurrent streams, supporting redundancy or migration.

---

## Session Lifecycle

A session represents an application-level connection between nodes.
It can be initiated locally or remotely and operates independently of which stream carries it.

Session states:

```
Routing → Open → Migrating → Closed
```

* **Routing** — negotiating session establishment.
* **Open** — active data exchange.
* **Migrating** — temporarily paused while moving to another stream.
* **Closed** — terminated gracefully or due to error.

Flow control ensures fairness and prevents buffer overruns through credit-based signaling (Read/Data).

---

## Frame Model 

Frames are the atomic units of communication over a stream.
Each frame carries a session nonce for routing.

Key frame types:

* **Query / Response** — session setup
* **Data / Read** — data transfer and flow control
* **Reset** — session termination
* **Ping / Pong** — stream liveness
* **Migrate** — signals session migration boundary

These minimal opcodes enable the multiplexing and lifecycle coordination of sessions without maintaining global state.

---

## Migration

Migration allows a session to seamlessly move from one active stream to another between the same peers.

* **Initiator** starts the process, designating a new target stream.
* **Responder** synchronizes upon receiving a migration marker (`Migrate` frame).
* Both sides temporarily pause writes while applying migration.
* After synchronization, data flow continues on the new stream.

Only the initiator sends the migration marker; the responder passively updates.
Migration signaling happens out-of-band, ensuring the old stream remains stable until the transition point.

Migration maintains session continuity during network transitions or stream replacements without disrupting the application layer.

---

## Peers Module and Integration

The `Peers` component coordinates all streams and sessions with a given identity.

Responsibilities:

* Maintain active stream and session registries.
* Dispatch inbound frames to corresponding sessions.
* Route outbound queries and responses.
* Emit link/unlink events when streams to a peer appear or disappear.

It acts as the central router between the underlying transports and higher Astral modules.

---

## Design Summary

**Separation of Concerns:**
Streams handle encryption, framing, and health; sessions handle logical continuity and application I/O.

**Migration Simplicity:**
Single-direction signaling and marker-based synchronization prevent race conditions while maintaining bidirectional data flow.

**Resilience:**
Multiple concurrent streams allow for redundancy and mobility across changing network conditions.

**Transparency:**
Applications using sessions are unaware of underlying stream changes; data continuity is preserved automatically.

**Extensibility:**
The modular design enables additional features like automatic migration triggers, multi-path routing, or quality-based stream selection in future versions.


### Current limitations and future improvements

- Operator must manually identify session and target stream nonces
- No automatic migration triggers (quality based stream selection, shutting down streams, etc.)
- Single stream failure terminates all sessions (no automatic recovery)
# Astral Nodes and Streams Overview

This page explains how identities, transports (exonet), secure links (Noise), frames, and the nodes module work together to establish long‑lived streams and route per‑query virtual connections.

It’s designed for contributors and operators who want to understand and tune link behavior (e.g., limiting network types, dealing with idle links, or inspecting active streams).

---

## Core components

- Exonet transports (mod/exonet)
  - Pluggable network dials (tcp, tor, utp, gw). Endpoints expose Network() and Address().
- Noise secure links (mod/nodes/src/noise, brontide/*)
  - Upgrades raw exonet.Conn to an authenticated, encrypted astral.Conn with LocalIdentity/RemoteIdentity.
- Frames (mod/nodes/src/frames)
  - A minimal binary framing protocol multiplexed over a secure conn (opcodes: Ping, Query, Response, Read, Data, Reset).
- Nodes stream (mod/nodes/src/stream.go)
  - Wraps frames.Stream, binds identities and endpoints, tracks liveness (ping) and lifecycle.
- Peers manager (mod/nodes/src/peers.go)
  - Holds current streams (Set[*Stream]) and per‑query virtual connections keyed by a Nonce.
- Virtual connections (mod/nodes/src/conn.go)
  - Per‑query Read/Write/Close split over frames, with credit‑based flow control.

---

## Big picture

```
+---------------------------+             +---------------------------+
|         Node A           |             |          Node B           |
|  - Identity: pubkey A    |             |   - Identity: pubkey B    |
|  - Router (RouteQuery)   |             |   - Router                |
|  - Nodes module          |             |   - Nodes module          |
|    • Peers: Set[*Stream] |<--frames--->|    • Peers: Set[*Stream]  |
|    • Conns: map[Nonce]*c |   (mux2)    |    • Conns: map[Nonce]*c  |
+------------^--------------+             +--------------^------------+
             |                                               |
             | secure link (Noise_XK, AEAD)                  |
             +--------------------+--------------------------+
                                  |
                    +-------------v-------------+
                    |     exonet transport      |
                    |   tcp | tor | utp | gw    |
                    +---------------------------+
```

- Exonet handles raw dialing by network. Noise authenticates peers and encrypts.
- Frames multiplex lightweight messages over the secure link; nodes.Stream wraps and manages health.
- Peers tracks streams and per‑query conns tied to a selected stream.

---

## Outbound link establishment (A → B)

```
ResolveEndpoints(B) --fan-in--> endpoints: exonet.Endpoint
         |
         v
  connectAtAny (N workers)
    for each endpoint e:
      Exonet.Dial(e) -> raw exonet.Conn
      Noise.HandshakeOutbound(A->B) -> secure astral.Conn
      Negotiate feature: "mux2"
      newStream(secure, outbound=true)
      addStream(stream)
        - streams.Add(stream)
        - log: "added out-stream with B (net)"
        - if first to B: emit EventLinked{B}
        - go readStreamFrames(stream)
        - (inbound only) go reflectStream()
```

- Workers race multiple endpoints; the first successful stream wins; others are closed as "excess stream".

---

## Inbound link acceptance (B ← A)

```
Exonet accept raw
  -> Noise.HandshakeInbound(B)
  -> Advertise features: ["mux2"]
  -> Read peer’s selection: "mux2" OK
  -> newStream(secure, outbound=false)
  -> addStream(stream)
```

- After admission, B is ready to receive queries from A.

---

## Query routing lifecycle

Outbound (A originates):
```
A.RouteQuery(ctx, q={Caller:A, Target:B, Query, Nonce})
  Peers.RouteQuery:
    - streams := all streams to B
    - if none: RouteNotFound
    - conn := new per-query virtual conn (state=Routing)
    - frame := frames.Query{Nonce, Query, Buffer=rbuf}
    - write Query on all streams to B
    - wait for Response or ctx.Done
      • if Accepted(Buffer): attach chosen stream, state=Open
        - spawn pump: app writes -> frames.Data (credit-based)
        - return conn (io.ReadWriteCloser)
      • if Rejected(code): delete conn, return ErrRejected
      • if ctx.Done: close conn, RouteNotFound
```

Inbound (B receives Query frame):
```
handleQuery(stream s, frames.Query f):
  - conn := new per-query conn {Nonce, stream=s, wsize=f.Buffer}
  - q := {Nonce, Caller=s.RemoteID, Target=s.LocalID, Query=f.Query}
  - q.Extra["origin"] = OriginNetwork
  - w, err := node.RouteQuery(ctx, q, conn)
    • if err: write Response{Nonce, ErrCode}
    • else:   write Response{Nonce, CodeAccepted, Buffer=rbuf}
              conn.state=Open
              spawn pump: service writes -> frames.Data
```

---

## Flow control & frames

- Data frames: up to 8192 bytes payload.
- Read frames: receiver signals bytes consumed to replenish sender credit (wsize).
- Reset frames: close a virtual connection.
- Default receive buffer per virtual connection: 4 MB.

```
Sender (A)                               Receiver (B)
--------------                           ----------------
wsize (credit) ---consume by Data---->   rbuf enqueue
                     <---Read(len)---    replenish wsize
```

State machine: Routing → Open → Closed. Overflow or invalid state triggers Reset/close.

---

## Stream health (pings) and teardown

- Stream.Write(non‑Ping) triggers a checker that sends Ping frames and expects Pong within pingTimeout (30s default).
- On timeout or error, the stream closes.
- On teardown:
  - streams.Remove(stream)
  - close all virtual conns attached to this stream
  - if this was the last stream to a remote identity, emit EventUnlinked{Remote}

---

## Events and observed endpoints

- EventLinked{NodeID}: emitted when the first stream to a remote identity appears.
- EventUnlinked{NodeID}: emitted when the last stream to that identity disappears.
- ObservedEndpoint reflection (inbound only): upon new inbound stream with a known RemoteEndpoint, B sends ObservedEndpointMessage{Endpoint} to A. Peers can use it as a candidate address (see ip_candidate_finder).

---

## Introspection

- List streams: `nodes.OpStreams` emits nodes.StreamInfo for each stream:
  - ID, LocalIdentity, RemoteIdentity, LocalEndpoint, RemoteEndpoint, Outbound flag (sorted by createdAt).
- Helpers in code:
  - isLinked(remote): true if any stream exists to identity.
  - linksTo(remote): returns []*Stream for identity grouping.

---

## Networks and endpoints

- Stream.Network() derives from exonet endpoints:
  - Typical values: "tcp", "tor", "utp", "gw".
- Stream.LocalEndpoint()/RemoteEndpoint() expose the underlying exonet endpoints when available.

---

## Defaults and safeguards

- Max Data payload per frame: 8192 bytes.
- Per‑query receive buffer: 4 MB default.
- Ping timeout: 30s; checker runs on demand after outbound activity.
- Outbound routing requires an existing link to Target; if none, RouteNotFound.
- Invalid frames, buffer overflow, duplicate nonces, or ping timeouts close the relevant connection.

---

## Key file references

- Nodes module (manager and routing):
  - mod/nodes/src/peers.go, mod/nodes/src/stream.go, mod/nodes/src/conn.go
- Frames:
  - mod/nodes/src/frames/*
- Noise handshake:
  - mod/nodes/src/noise/*, brontide/*
- Exonet transports:
  - mod/exonet/src/*, mod/tcp/*, mod/tor/*, mod/utp/*, mod/gateway/*
- Ops and objects:
  - mod/nodes/src/op_streams.go, mod/nodes/events.go, mod/nodes/observed_endpoint_message.go

---

## Glossary

- Link: An authenticated, encrypted connection between two identities (a stream with negotiated features).
- Stream: The mux2 framing layer over a secure link; carries frames.
- Virtual connection: A per‑query, nonce‑keyed logical pipe over a stream.
- Endpoint: A transport address (network + address), e.g., tcp:1.2.3.4:1234 or tor:digest:port.


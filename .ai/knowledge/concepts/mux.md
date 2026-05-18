# Session Multiplexer

## Model

A **Link** is the physical carrier: one brontide-encrypted connection over one transport
(`tcp`, `kcp`, `tor`, `gw`). Establishment is expensive: Noise XK handshake, possible NAT
traversal, possible Tor circuit.

A **Session** is a logical pipe over a Link: one routed Query, tracked by a 64-bit
`Nonce`, with independent read/write buffers and flow control. Sessions are cheap; one
Link can carry hundreds.

The mux layer sits between the two:

```
Session (logical, many)
  └─ frames/mux (multiplexes over one Link)
       └─ Link (physical, expensive)
```

Opening 100 services to the same peer uses one brontide handshake, not 100.

## Multiplexing Invariants

- **Cost**: reuse one Link per identity to amortize handshake, NAT hole-punch, and Tor
  setup cost.
- **Symmetry**: after Link establishment, both peers use the Link identically. The mux
  layer has no per-session initiator/responder role. Either peer may send a Query frame.
- **Isolation**: `Nonce` identifies the session. Flow control, buffer state, and
  close/reset are per-session, so one slow reader does not block unrelated sessions.

## Flow control

Each session has independent credit.

- Sender tracks `wsize`, the remaining remote buffer.
- `Write` blocks when `wsize == 0`.
- Receiver sends a `Read` frame after draining bytes from its buffer.
- `Read` frames increment `wsize`; `defaultBufferSize` is 4 MB.

Trust boundary: sender credit is controlled by the receiver. A fast sender cannot
overflow a slow receiver, and a blocked session does not stall other sessions on the same
Link.

During `stateMigrating`, `Write` blocks until migration completes. Data is not lost or
reordered across the carrier switch.

## Session migration

A session can migrate from one Link to a better Link while it remains open and in-flight.
The current Link's pressure detector decides when to upgrade: sustained throughput above
a threshold or RTT above the transport baseline triggers `connectivityUpgrade`.

Upgrade selection:

1. Prefer an existing lower-pressure Link to the same peer.
2. If none exists, attempt NAT traversal and establish a fresh Link.

### Preserved State

- Session `Nonce`; the application-level pipe identity does not change.
- Buffered unread data already pushed into the read buffer; receiver reads continue.
- Query string, remote identity, and total byte counter.

### Replaced State

- Physical carrier; the session detaches from the old Link and attaches to the new one
  with `c.link = c.migratingTo`.
- Unsent writes on the old Link; `Write` blocks in `stateMigrating` until
  `CompleteMigration` switches the Link and reopens the state.

## Migration Protocol

Both peers coordinate migration:

1. Initiator calls `Migrate`, entering `stateMigrating`; `Write` now blocks.
2. Initiator sends a `Migrate` frame on the old Link.
3. Initiator waits for the responder's echoed `Migrate` frame, confirming all old-link
   data has been processed.
4. Responder receives the `Migrate` frame, echoes one back, and signals
   `OpMigrateSession`.
5. After both sides exchange `Migrate` frames and the initiator confirms it drained the
   old Link, `CompleteMigration` switches `c.link` to the new Link, returns state to
   `stateOpen`, and unblocks `Write`.

The session stays open. The application-facing `io.ReadWriter` keeps working over the new
carrier.

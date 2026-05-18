# Transport

* Transport modules register under `exonet`.
* Supported transports: TCP, KCP/UDP, Tor, and gateway relay.
* Upper layers are transport-agnostic.

## Exonet

* `exonet` is the pluggable transport registry.
* `Endpoint` is a transport name plus an address string.
* `Endpoint` is a serializable Object.
* `exonet.Conn` is a raw unauthenticated byte stream with endpoint metadata.
* Transport modules register a Dialer, Parser, and Unpacker.

## Layer Stack

```
Session          — one routed Query, flow-controlled
  └─ frames/mux  — multiplexes Sessions over one Link
       └─ brontide — Noise XK; secp256k1 auth; ChaCha20-Poly1305 encryption
            └─ exonet.Conn — raw transport bytes (tcp / kcp / tor / gw)
```

## Transports

| Transport | Protocol | Use |
|---|---|---|
| `tcp` | TCP/IP | direct, stable IPs |
| `kcp` | KCP/UDP | NAT traversal after hole-punch |
| `tor` | Tor hidden services | anonymity |
| `gw` | gateway relay | unreachable nodes |

## Link Strategies

* Race strategies in parallel.
* The first successful brontide handshake wins.

| Strategy | Approach |
|---|---|
| Basic | Resolve published endpoints, dial directly |
| NAT | Coordinate UDP hole-punch via `nat`, then dial KCP |
| Tor | Dial Tor hidden service endpoint |

# uTP Module

Simple uTP (Micro Transport Protocol) transport support for Astral.

## What is uTP?
uTP is a lightweight, reliable protocol that runs over UDP. Think of it like "TCP over UDP" but tuned to be friendly to the network:
- Reliable + in-order delivery
- Built-in congestion control (aims to not saturate links)
- Often better across NAT and variable links
- Used widely in BitTorrent

## What this module does
- Adds support for `utp:` endpoints in Astral
- Lets nodes listen and connect using uTP
- Parses endpoint strings like `utp:host:port`
- Integrates with the existing endpoint resolver + dialing system

## Endpoint format
```
utp:<host>:<port>
```
Examples:
- `utp:127.0.0.1:7777`
- `utp:example.org:6000`
- IPv6: `utp:[2001:db8::1]:9000`

## Add a uTP endpoint to a node
Use `astral-query` to attach a new endpoint to a known identity:
```
astral-query localnode:nodes.add_endpoint -id {identity} -endpoint utp:{address}:{port}
```
Example:
```
astral-query localnode:nodes.add_endpoint -id zK3...AbC -endpoint utp:203.0.113.10:7777
```

## Testing connectivity
1. Add your uTP endpoint (as above)
2. From another node, perform a query or connect normally—the resolver will use the uTP address

## Notes
- Requires UDP reachability for best results
- If behind strict NAT, port forwarding may still help
- Keep ports stable so other nodes can reconnect

## Removal
Remove or replace like any other endpoint via the node endpoint management commands.

---
Minimal, clear, and fast: that's the goal of this module.


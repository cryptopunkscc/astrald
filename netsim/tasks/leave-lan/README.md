# leave-lan

Makes `--vm` (node2) **leave the LAN** with respect to `--peer` (node1). Two host-side
steps:

1. **Seed the peer with the leaver's onion while the LAN is still up.** Once the LAN is
   cut, node1 could no longer ask node2 for its address, so `leave-lan` first records
   node2's `.onion` on node1 (`<vm>:nodes.resolve_endpoints` → `nodes.add_endpoint`),
   keeping node2 reachable over Tor.
2. **Sever the LAN.** An nftables drop (a dedicated `netsimcut` table) blackholes all
   traffic to/from the peer's LAN (`10.77.0.0/24`) address. The NIC and Internet/Tor
   egress (the WAN NAT) stay up — only the direct LAN path is cut.

astrald has no link keepalive, so the dead LAN link lingers as a (blackholed) stale
entry rather than closing — which is why the agent (`link-over-tor`) must *force* the
Tor link rather than wait for an automatic reconnect. `verify.py` independently confirms
`--vm` can no longer open a TCP connection to the peer's astral port (1791) — only a
connect **timeout** counts (a refusal/reset would be inconclusive, not a pass).
Host-driven. Used by `tor-link.story` after `enable-tor`.

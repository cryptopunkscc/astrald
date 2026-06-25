# leave-lan

On the host, seeds `--peer` (node1) with `--vm` (node2)'s onion (`nodes.resolve_endpoints` → `nodes.add_endpoint`), then nftables-drops the LAN path between them, leaving node2 reachable from node1 only over Tor. verify.py asserts node2 can no longer TCP-connect to node1's LAN address on port 1791.

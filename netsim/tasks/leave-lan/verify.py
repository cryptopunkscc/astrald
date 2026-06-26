#!/usr/bin/env python3
"""verify leave-lan: <vm> has withdrawn its LAN identity, so it genuinely left the
10.77 LAN (not merely had its packets filtered).

The cut (run.sh) flushes <vm>'s own 10.77 address, which is what astrald observes as
"left the network": it polls net.InterfaceAddrs() every 3s and advertises one tcp
endpoint per assigned IP, so removing the address fires EventNetworkAddressChanged and
withdraws the 10.77 tcp endpoint. (A packet-filter DROP -- or a bare link/carrier down --
leaves the IPv4 address in place and is invisible to that monitor.)

This is a blind, deterministic host-side check: it reads <vm>'s own network state over
ssh, independent of astral, and asserts two consequences of the address withdrawal --
<vm> has (1) no 10.77 LAN address and (2) no route into the 10.77 subnet. Neither depends
on a TCP probe's error code, which would vary with the WAN default route (a connect to
the LAN falls through to the WAN NAT and times out rather than returning ENETUNREACH).
astrald's reaction -- re-linking over Tor -- is asserted separately by link-over-tor.
"""
import argparse
import os
import sys

# why: realpath crosses netsim's per-task symlink to reach the sibling tasks/_lib
sys.path.insert(0, os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "_lib"))
import astralapi  # noqa: E402


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", default="node2")      # the node that left the LAN
    ap.add_argument("--peer", default="node1")    # the node it can no longer reach
    args, _ = ap.parse_known_args()

    # 1) the leaver no longer holds any 10.77 LAN address (the thing astrald keys on)
    lan_ip = astralapi.peer_lan_ip(args.vm)
    # 2) and has no route into the 10.77 subnet (the connected route went with the address)
    lan_routes = [ln for ln in (astralapi.ssh(args.vm, "ip -o route show") or "").splitlines()
                  if "10.77." in ln]

    if lan_ip:
        sys.stderr.write(f"leave-lan verify FAILED: {args.vm} still holds a LAN address "
                         f"({lan_ip}) -- it has not left the 10.77 LAN.\n")
        return 1
    if lan_routes:
        sys.stderr.write(f"leave-lan verify FAILED: {args.vm} still has a route into the "
                         "10.77 LAN:\n  " + "\n  ".join(lan_routes) + "\n")
        return 1

    print(f"leave-lan OK: {args.vm} withdrew its 10.77 LAN address and route -- it has left "
          f"the LAN (astrald re-links to {args.peer} over Tor; asserted by link-over-tor).")
    return 0


if __name__ == "__main__":
    sys.exit(main())

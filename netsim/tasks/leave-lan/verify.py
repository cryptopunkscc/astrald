#!/usr/bin/env python3
"""verify leave-lan: <vm> can no longer reach <peer> over the LAN.

Independent host-side check: from <vm>, a TCP connect to the peer's LAN address on
the astral port (1791) must NOT succeed (the nftables drop blackholes it ->
timeout). The peer's LAN IP is resolved from the peer. No astral-query here -- this
is a raw socket probe, run through tasks/_lib/astralapi.py's ssh transport.
"""
import argparse
import os
import sys

# why: realpath crosses netsim's per-task symlink to reach the sibling tasks/_lib
sys.path.insert(0, os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "_lib"))
import astralapi  # noqa: E402

PORT = 1791


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", default="node2")      # the node that left the LAN
    ap.add_argument("--peer", default="node1")    # the node it can no longer reach
    args, _ = ap.parse_known_args()

    ip = astralapi.peer_lan_ip(args.peer)
    if not ip:
        sys.stderr.write(f"leave-lan verify FAILED: could not resolve {args.peer}'s 10.77 LAN IP.\n")
        return 1

    # Only a TIMEOUT proves the nftables DROP blackholed the path. A connect that
    # succeeds means the LAN is not severed; a refusal/reset (or any other error) means
    # the path is reachable but the port is closed for another reason -> inconclusive,
    # NOT a pass (would otherwise false-pass if the drop rule were missing).
    probe = (
        "python3 -c 'import socket\n"
        "s=socket.socket(); s.settimeout(3)\n"
        f"try:\n s.connect((\"{ip}\",{PORT})); print(\"open\")\n"
        "except socket.timeout:\n print(\"timeout\")\n"
        "except Exception as e:\n print(\"err:\"+type(e).__name__)'"
    )
    result = (astralapi.ssh(args.vm, probe) or "").strip()

    if result == "timeout":
        print(f"leave-lan OK: {args.vm} can no longer reach {args.peer} ({ip}:{PORT}) over the LAN "
              "(connect times out -- blackholed)")
        return 0

    if result == "open":
        sys.stderr.write(f"leave-lan verify FAILED: {args.vm} still reaches {args.peer} "
                         f"({ip}:{PORT}) over the LAN (connect succeeded).\n")
    else:
        sys.stderr.write(f"leave-lan verify FAILED: probe to {args.peer} ({ip}:{PORT}) was "
                         f"inconclusive ({result!r}) -- expected a timeout from the drop, not a "
                         "refusal/reset.\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

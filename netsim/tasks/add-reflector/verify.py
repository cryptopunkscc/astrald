#!/usr/bin/env python3
"""verify add-reflector: both NAT'd peers learned their public endpoint by reflection,
arming astrald's nat module.

After enter-nat hides each peer behind a symmetric masquerade NAT, the peer can't see its
own public address; add-reflector wires a public reflector that observes the peer's SNAT'd
source and reflects it back. The observable result -- and exactly what flips astrald's nat
module on (`evaluateEnabled`: the `enabled` setting defaults on AND
len(PublicIPCandidates())>0, `mod/nat/src/module.go`) -- is that each peer's public IP
candidates now include its TEST-NET alias 198.51.100.<lan-octet>.

Blind host-side check: for each peer, derive its 198.51.100.<octet> from its 10.77 LAN
octet, then query the peer's astrald `ip.public_ip_candidates` and assert that address is
present. That address being a public candidate == the peer is armed (nat can/does enable).

Note: the peer's astrald runs INSIDE netns "priv" (enter-nat), so its WS apphost port is
netns-local and NOT reachable over the ssh -L forward -- we query via the Go `astral-query`
CLI over the apphost unix socket (shared mount ns), which crosses the net-ns boundary. So
this verify uses astralapi.ssh directly, not the astral-py client.
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
    ap.add_argument("--reflector", default="reflector")     # accepted (same argv as run.sh), unused here
    ap.add_argument("--vm", dest="vms", action="append", default=[])
    args, _ = ap.parse_known_args()
    peers = args.vms or ["node1", "node2"]

    failed = []
    for p in peers:
        lan = astralapi.peer_lan_ip(p)                      # e.g. "10.77.0.12"
        if not lan:
            failed.append(f"{p}: could not read its 10.77 LAN address")
            continue
        want = "198.51.100." + lan.split(".")[-1]           # the peer's public TEST-NET alias
        # local introspection op, ungated; astrald is in netns priv -> unix-socket CLI.
        raw = astralapi.ssh(p, "astral-query ip.public_ip_candidates -out json") or ""
        if want in raw:
            print(f"add-reflector OK: {p} nat armed -- public candidate {want} present.")
        else:
            failed.append(f"{p}: public candidate {want} NOT among ip.public_ip_candidates")
            sys.stderr.write(f"  {p} ip.public_ip_candidates:\n{raw or '(empty)'}\n")

    if failed:
        for f in failed:
            sys.stderr.write(f"add-reflector verify FAILED: {f}\n")
        return 1
    print(f"add-reflector verified: nat armed on all peers ({', '.join(peers)})")
    return 0


if __name__ == "__main__":
    sys.exit(main())

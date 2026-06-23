#!/usr/bin/env python3
"""verify link-over-tor: node1 holds a live link to the peer over Tor.

Independent host-side check (does not trust the agent): nodes.links on node1 must list a
link whose Network is "tor". (We assert the transport, not the .onion endpoint string —
an inbound tor link legitimately has no remote onion, so requiring ".onion" would
false-negative; node2 is the only sibling, so a tor link is a tor link to node2.) Also
cross-checks the agent's record. Reaches the VM via netsim ssh.
"""
import argparse
import json
import subprocess
import sys


def ssh(vm, remote):
    p = subprocess.run(["netsim", "ssh", vm, "--", remote], capture_output=True, text=True)
    return p.stdout


def jfile(vm, name):
    try:
        return json.loads(ssh(vm, f"cat /home/tester/{name}") or "{}") or {}
    except json.JSONDecodeError:
        return {}


def ep_addr(e):
    """Address string of an exonet.Endpoint, whether it marshals bare or as {Type,Object}."""
    if isinstance(e, str):
        return e
    if isinstance(e, dict):
        o = e.get("Object")
        return o if isinstance(o, str) else ""
    return ""


def tor_links(stream):
    """(RemoteIdentity, endpoint-address) for links whose Network == 'tor'."""
    out = []
    for ln in (stream or "").splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            o = json.loads(ln)
        except json.JSONDecodeError:
            continue
        ob = o.get("Object") or {}
        if str(ob.get("Network")) == "tor":
            out.append((str(ob.get("RemoteIdentity", "")), ep_addr(ob.get("RemoteEndpoint"))))
    return out


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", default="node1")      # the operator; records tor.json here
    ap.add_argument("--peer", default="node2")    # the node that left the LAN
    args, _ = ap.parse_known_args()

    tor = jfile(args.vm, "tor.json")               # agent: peer_onion, link_network
    net = str(tor.get("link_network", ""))
    onion = str(tor.get("peer_onion", ""))

    # Decisive: an actual link over Tor from node1 (to the only sibling, the peer).
    links = tor_links(ssh(args.vm, "astral-query nodes.links -out json"))

    notes = []
    if net != "tor":
        notes.append(f"agent recorded link_network={net!r} (expected 'tor')")
    if not onion:
        notes.append("agent recorded no peer_onion")

    if links:
        ep = links[0][1] or "(inbound, no remote onion)"
        print(f"link-over-tor OK: {args.vm} holds a link to {args.peer} over Tor (endpoint {ep}).")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    sys.stderr.write(f"link-over-tor verify FAILED: {args.vm} has no link to {args.peer} over Tor.\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    sys.stderr.write(f"  nodes.links:\n{ssh(args.vm, 'astral-query nodes.links -out json')}\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

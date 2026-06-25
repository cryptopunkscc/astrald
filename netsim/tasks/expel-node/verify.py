#!/usr/bin/env python3
"""verify expel-node: node1 (the User) permanently banned node2 from the swarm.

Independent check (does not trust run.sh); reaches the VMs via netsim ssh. Asserts node2
is recorded in user.list_expelled and is gone from node1's user.swarm_status roster
(user.OpSwarmStatus -> ActiveNodes filters the expelledSet). Link state is not asserted.

node2's identity comes from node1's siblings.json (recorded by adopt-node), NOT from node2
itself: once expelled, node2 rejects user.info (query rejected (2) untokened, auth_failed
with the User token — it no longer accepts the User it was banned from), so it is not a
usable identity source.
"""
import argparse
import json
import subprocess
import sys


def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout."""
    p = subprocess.run(["netsim", "ssh", vm, "--", remote],
                       capture_output=True, text=True)
    return p.stdout


def jfile(vm, name):
    """/home/tester/<name> on the VM, parsed as a dict."""
    try:
        return json.loads(ssh(vm, f"cat /home/tester/{name}") or "{}") or {}
    except json.JSONDecodeError:
        return {}


def objs(stream):
    """astral-query -out json emits one object per line + an eos terminator."""
    out = []
    for ln in (stream or "").splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            out.append(json.loads(ln))
        except json.JSONDecodeError:
            pass
    return out


def swarm_identities(stream):
    """Set of node identities listed in a user.swarm_status stream."""
    ids = set()
    for o in objs(stream):
        ob = o.get("Object")
        if isinstance(ob, dict) and ob.get("Identity"):
            ids.add(ob["Identity"])
    return ids


def contains_identity(value, ident):
    """True if `ident` appears anywhere in a parsed JSON value (string match)."""
    if isinstance(value, str):
        return value == ident
    if isinstance(value, dict):
        return any(contains_identity(v, ident) for v in value.values())
    if isinstance(value, list):
        return any(contains_identity(v, ident) for v in value)
    return False


def is_expelled(stream, ident):
    """True if a user.list_expelled stream bans `ident` (as a SignedExpulsion Subject)."""
    for o in objs(stream):
        if contains_identity(o.get("Object", o), ident):
            return True
    return False


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--node1", default="node1")
    ap.add_argument("--node2", default="node2")
    args, _ = ap.parse_known_args()
    vm1, vm2 = args.node1, args.node2

    # node1 acts as the User (token from bootstrap); list_expelled / swarm_status require
    # the caller to be the contract issuer, so they run under that token.
    info1 = jfile(vm1, "user.json")
    U = "".join(str(info1.get("user_id", "")).split())
    TOKEN = f"export ASTRALD_APPHOST_TOKEN={info1.get('user_token', '')};"

    # node2's identity from node1's siblings.json (recorded by adopt-node) — a stable
    # source. The expelled node itself can't be queried (post-ban node2 rejects user.info).
    sibs = jfile(vm1, "siblings.json")
    sib_ids = ["".join(str(x).split()) for x in (sibs.get("sibling_ids") or []) if x]
    s2 = sib_ids[0] if sib_ids else None

    n1_expelled = ssh(vm1, TOKEN + " astral-query user.list_expelled -out json")
    n1_swarm = ssh(vm1, TOKEN + " astral-query user.swarm_status -out json")
    members = swarm_identities(n1_swarm)

    errs = []
    if not U:
        errs.append("no user_id in node1's user.json")
    if not s2:
        errs.append("no sibling_ids in node1's siblings.json — can't identify the expelled node")
    if s2 and not is_expelled(n1_expelled, s2):
        errs.append(f"node2 {s2} is NOT in node1's user.list_expelled "
                    "(expulsion was never issued — agent did not expel the node)")
    if s2 and s2 in members:
        errs.append(f"node2 {s2} still appears in node1's user.swarm_status "
                    "(roster not reduced — expelledSet filter did not drop it)")

    if errs:
        sys.stderr.write("expel-node verify FAILED:\n")
        for e in errs:
            sys.stderr.write(f"  - {e}\n")
        return 1

    print(f"expel OK: User {U[:8]}.. banned node2 {s2[:8]}.. — recorded in "
          f"user.list_expelled and dropped from user.swarm_status "
          f"({len(members)} member(s) remain).")
    return 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""verify expel-node: node1 (the User) permanently banned node2 from the swarm.

Independent both-ends check (does not trust run.sh); reaches the VMs via netsim ssh.
The core property — confirmed in code (user.OpSwarmStatus -> ActiveNodes filters the
expelledSet) — is that an expelled node yields FEWER swarm_status results: node2 is
gone from node1's roster, recorded in user.list_expelled, and the link is torn down.
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


def info(vm):
    """The agent's $HOME/info.json (/home/tester/info.json) on the VM, as a dict."""
    try:
        return json.loads(ssh(vm, "cat /home/tester/info.json") or "{}") or {}
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


def contract(stream):
    """(Issuer, Subject) of the active contract from a user.info stream."""
    for o in objs(stream):
        ob = o.get("Object")
        if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
            c = ob["Contract"].get("Contract", {})
            return c.get("Issuer"), c.get("Subject")
    return None, None


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


def has_link_to(links, identity):
    """True if a nodes.links stream contains an active link to `identity`."""
    for o in objs(links):
        ob = o.get("Object")
        if isinstance(ob, dict) and ob.get("RemoteIdentity") == identity:
            return True
    return False


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--node1", default="node1")
    ap.add_argument("--node2", default="node2")
    args, _ = ap.parse_known_args()
    vm1, vm2 = args.node1, args.node2

    # node1 acts as the User (token from bootstrap); the expel op requires the caller
    # to be the contract issuer, so list_expelled / swarm_status run under that token.
    info1 = info(vm1)
    U = "".join(str(info1.get("user_id", "")).split())
    TOKEN = f"export ASTRALD_APPHOST_TOKEN={info1.get('user_token', '')};"

    n1_info = ssh(vm1, TOKEN + " astral-query user.info -out json")
    n1_swarm = ssh(vm1, TOKEN + " astral-query user.swarm_status -out json")
    n1_expelled = ssh(vm1, TOKEN + " astral-query user.list_expelled -out json")
    n1_links = ssh(vm1, "astral-query nodes.links -out json")

    # node2 still holds its membership contract (expel bans, it does not revoke the
    # contract), so its identity is still readable from its own user.info.
    n2_info = ssh(vm2, "astral-query user.info -out json")
    n2_links = ssh(vm2, "astral-query nodes.links -out json")

    _, s1 = contract(n1_info)   # node1's identity (the swarm User's own node)
    _, s2 = contract(n2_info)   # node2's identity (the expelled subject)
    members = swarm_identities(n1_swarm)

    errs = []
    if not U:
        errs.append("no user_id in node1's info.json")
    if not s2:
        errs.append("could not resolve node2's identity from its user.info")
    if not is_expelled(n1_expelled, s2):
        errs.append(f"node2 {s2} is NOT in node1's user.list_expelled "
                    "(expulsion was never issued — agent did not expel the node)")
    if s2 and s2 in members:
        errs.append(f"node2 {s2} still appears in node1's user.swarm_status "
                    "(roster not reduced — expelledSet filter did not drop it)")
    if s2 and has_link_to(n1_links, s2):
        errs.append(f"node1 still holds an active link to expelled node2 {s2} "
                    "(applyExpulsion did not close the link)")
    if s1 and has_link_to(n2_links, s1):
        errs.append(f"node2 still holds an active link back to node1 {s1} "
                    "(link not torn down on the peer end)")

    if errs:
        sys.stderr.write("expel-node verify FAILED:\n")
        for e in errs:
            sys.stderr.write(f"  - {e}\n")
        return 1

    print(f"expel OK: User {U[:8]}.. banned node2 {s2[:8]}.. — recorded in "
          f"user.list_expelled, dropped from user.swarm_status ({len(members)} "
          f"member(s) remain), and the link is torn down on both ends")
    return 0


if __name__ == "__main__":
    sys.exit(main())

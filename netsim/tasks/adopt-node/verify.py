#!/usr/bin/env python3
"""verify adopt-node: node1 and node2 linked into one User swarm, symmetric roster.

Independent both-ends check (does not trust run.sh); reaches the VMs via netsim ssh.
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


def contract(info):
    """(Issuer, Subject) of the active contract from a user.info stream."""
    for o in objs(info):
        ob = o.get("Object")
        if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
            c = ob["Contract"].get("Contract", {})
            return c.get("Issuer"), c.get("Subject")
    return None, None


def linked_sibling(swarm):
    """Identity of the first Linked sibling in a user.swarm_status stream."""
    for o in objs(swarm):
        ob = o.get("Object")
        if isinstance(ob, dict) and ob.get("Linked"):
            return ob.get("Identity")
    return None


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

    # node1 acts as the User (token from bootstrap-user-software-key); node2 answers under its
    # node identity (it holds the contract after the adoption).
    info1 = info(vm1)
    U = "".join(str(info1.get("user_id", "")).split())
    TOKEN = f"export ASTRALD_APPHOST_TOKEN={info1.get('user_token', '')};"
    n1_info = ssh(vm1, TOKEN + " astral-query user.info -out json")
    n1_swarm = ssh(vm1, TOKEN + " astral-query user.swarm_status -out json")
    n2_info = ssh(vm2, "astral-query user.info -out json")
    n2_links = ssh(vm2, "astral-query nodes.links -out json")
    # node2's own swarm view: swarm_status derives from node2's active contract,
    # not the caller, so no token is needed; post-#348 it must list node1 too.
    n2_swarm = ssh(vm2, "astral-query user.swarm_status -out json")

    i1, s1 = contract(n1_info)
    i2, s2 = contract(n2_info)
    sib = linked_sibling(n1_swarm)
    n2_sib = linked_sibling(n2_swarm)
    linkback = has_link_to(n2_links, s1)

    errs = []
    if not U:
        errs.append("no user_id in node1's info.json")
    if i1 != U:
        errs.append(f"node1 contract issuer {i1} != User {U}")
    if i2 != U:
        errs.append(f"node2 contract issuer {i2} != User {U} (node2 not adopted under this User)")
    if not s1:
        errs.append("node1 has no active contract subject")
    if not s2:
        errs.append("node2 has no active contract subject")
    if s2 and sib != s2:
        errs.append(f"node1's linked sibling {sib} != node2 {s2}")
    if s1 and n2_sib != s1:
        errs.append(f"node2's linked sibling {n2_sib} != node1 {s1} "
                    "(node2 does not list node1 -- swarm roster not symmetric; #348 regression?)")
    if not linkback:
        errs.append(f"node2 has no active link back to node1 ({s1})")

    if errs:
        sys.stderr.write("adopt-node verify FAILED:\n")
        for e in errs:
            sys.stderr.write(f"  - {e}\n")
        return 1

    print(f"swarm OK: User {U[:8]}.. ; node1 {s1[:8]}.. <-link-> node2 {s2[:8]}.. ; "
          f"both under one User; each lists the other as a Linked sibling (symmetric roster)")
    return 0


if __name__ == "__main__":
    sys.exit(main())

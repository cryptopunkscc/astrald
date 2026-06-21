#!/usr/bin/env python3
"""verify link-swarm: node1 and node2 must be linked into one User swarm.

INDEPENDENT both-ends check -- it does not trust run.sh. It pulls raw JSON from
both nodes and asserts THREE facts on the host; together they prove the swarm from
both ends:
  1. both nodes hold an active contract issued by the SAME User
     (user.info: Issuer == the bootstrap User on each; Subject == that node);
  2. node1, acting as the User, lists node2 as a Linked sibling (user.swarm_status);
  3. a mutual authenticated link exists (node2 nodes.links -> node1).

Runs on the host (invoked by the verify.sh shim); reaches the VMs with `netsim ssh`.

NOTE on "routed query": an earlier plan probed `<peer>:.spec` as the proof. That is
NOT valid -- node introspection ops (.spec/.id/.ping) are served locally and do not
route to a sibling by node-id, so they fail even on a fully formed swarm. The
contract + link + sibling triple above is the real proof.

astral-query ... -out json emits a JSON *stream* (one object per line, then an
{"Type":"eos"} terminator), so everything is parsed line-by-line, not as one doc.
"""
import argparse
import json
import subprocess
import sys

TOKEN = "export ASTRALD_APPHOST_TOKEN=$(cat /home/tester/.netsim/user.token);"


def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout."""
    p = subprocess.run(["netsim", "ssh", vm, "--", remote],
                       capture_output=True, text=True)
    return p.stdout


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

    # node1 acts as the User (token from bootstrap-user); node2 answers under its
    # node identity (it holds the contract after the adoption).
    U = "".join(ssh(vm1, "cat /home/tester/.netsim/user.id").split())
    n1_info = ssh(vm1, TOKEN + " astral-query user.info -out json")
    n1_swarm = ssh(vm1, TOKEN + " astral-query user.swarm_status -out json")
    n2_info = ssh(vm2, "astral-query user.info -out json")
    n2_links = ssh(vm2, "astral-query nodes.links -out json")

    i1, s1 = contract(n1_info)
    i2, s2 = contract(n2_info)
    sib = linked_sibling(n1_swarm)
    linkback = has_link_to(n2_links, s1)

    errs = []
    if not U:
        errs.append("no User id recorded on node1 (~/.netsim/user.id)")
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
    if not linkback:
        errs.append(f"node2 has no active link back to node1 ({s1})")

    if errs:
        sys.stderr.write("link-swarm verify FAILED:\n")
        for e in errs:
            sys.stderr.write(f"  - {e}\n")
        return 1

    print(f"swarm OK: User {U[:8]}.. ; node1 {s1[:8]}.. <-link-> node2 {s2[:8]}.. ; "
          f"both under one User; node1 lists node2 as Linked sibling")
    return 0


if __name__ == "__main__":
    sys.exit(main())

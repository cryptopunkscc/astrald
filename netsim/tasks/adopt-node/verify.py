#!/usr/bin/env python3
"""verify adopt-node: node1 and node2 linked into one User swarm, symmetric roster.

Independent both-ends check (does not trust run.sh). Queries reach each VM's
apphost through the shared astral-py client (tasks/_lib/netsim_astral.py), which
forwards to the lockstep Go astral-query CLI for any op it can't serve.
"""
import argparse
import os
import sys

# why: realpath crosses netsim's per-task symlink to reach the sibling tasks/_lib
sys.path.insert(0, os.path.join(
    os.path.dirname(os.path.dirname(os.path.realpath(__file__))), "_lib"))
import netsim_astral as na  # noqa: E402


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--node1", default="node1")
    ap.add_argument("--node2", default="node2")
    args, _ = ap.parse_known_args()
    vm1, vm2 = args.node1, args.node2

    info1 = na.home_json(vm1, "user.json")
    siblings = na.home_json(vm1, "siblings.json")  # adopt-node agent: swarm sibling ids
    sib_ids = ["".join(str(x).split()) for x in (siblings.get("sibling_ids") or []) if x]
    U = "".join(str(info1.get("user_id", "")).split())
    token = info1.get("user_token", "")

    # node1 acts as the User (token from bootstrap-user-software-key); node2 answers
    # under its node identity (it holds the contract after the adoption).
    with na.connect(vm1, token=token) as n1:
        i1, s1 = na.contract(n1.call("user.info"))
        sib = na.linked_sibling(n1.call("user.swarm_status"))
    # node2's own swarm view: swarm_status derives from node2's active contract,
    # not the caller, so no token is needed; post-#348 it must list node1 too.
    with na.connect(vm2) as n2:
        i2, s2 = na.contract(n2.call("user.info"))
        linkback = na.has_link_to(n2.call("nodes.links"), s1)
        n2_sib = na.linked_sibling(n2.call("user.swarm_status"))

    errs = []
    if not U:
        errs.append("no user_id in node1's user.json")
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
    if not sib_ids:
        errs.append("node1 recorded no sibling_ids in ~/siblings.json")
    elif s2 and s2 not in sib_ids:
        errs.append(f"node1's recorded sibling_ids {sib_ids} do not include adopted node {s2}")

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

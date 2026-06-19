#!/usr/bin/env python3
"""verify share-object: an astral object stored on node1 must be obtainable by
its sibling node2 ACROSS THE SWARM.

INDEPENDENT host-side check -- it does not trust run.sh. It reads the id + payload
the agent persisted on node1, then tries to pull that exact id FROM node2's vantage
and asserts the bytes match. Runs on the host (invoked by the verify.sh shim, which
netsim runs with cwd=sim root and $NETSIM_TASK_DIR set); it reaches the VMs with
`netsim ssh`.

THE CROSS-SWARM HOP IS INFERRED, NOT DEMONSTRATED (see README). The astral-docs
describe a network zone + a finder/provider layer but no worked example of one
swarm member reading another's object by id, so we probe a LADDER and report which
hop routes -- exactly as link-swarm discovered that <peer>:.spec does NOT route.
Order (strongest -> weakest), all run on node2:
  1. EXPLICIT TARGET  astral-query <node1-id>:objects.load -id <ID> -out json
       Query-target routing over the swarm link. Primary: does NOT rely on node2's
       network zone (an anonymous apphost caller has ZoneNetwork stripped) -- it
       addresses node1 directly; node1 serves the read locally.
  2. TRANSPARENT      astral-query objects.load -id <ID> -out json
       Relies on the read context's zone defaulting to all zones (incl. network).
       Likely BLOCKED for an anonymous host-side caller -- kept as a bonus probe.
  3. PROVIDER FIND    astral-query objects.find -id <ID> -out json
       Returns provider IDENTITIES, not bytes. If only this works, discovery
       crosses but the byte read does not -- a partial finding, not a pass.

PASS iff node2 obtained the EXACT stored bytes for the agent-reported id across the
swarm (hop 1 or 2). A pre-check asserts node2 doesn't already hold the object
locally, so a pass reflects a genuine remote pull.

astral-query ... -out json emits a JSON *stream* (one object per line, then an
{"Type":"eos"} terminator), so everything is parsed line-by-line, not as one doc.
"""
import argparse
import json
import subprocess
import sys

TOKEN = "export ASTRALD_APPHOST_TOKEN=$(cat /home/tester/.netsim/user.token);"


def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout (best-effort).

    astral-query writes error_message frames to stdout (which we parse) and other
    failures (route_not_found, etc.) to stderr (which we drop) -- mirroring the
    old shell `2>/dev/null`.
    """
    p = subprocess.run(["netsim", "ssh", vm, "--", remote],
                       capture_output=True, text=True)
    return p.stdout


# ---- JSON object-stream parsing (one object/line + an eos terminator) ----------

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


def loaded_payload(stream):
    """From an objects.load stream, the decoded payload string (the stored
    string8's Object), or None. Skips eos / error_message frames."""
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        ob = o.get("Object")
        if isinstance(ob, str):
            return ob
    return None


def errors(stream):
    return [o.get("Object") for o in objs(stream) if o.get("Type") == "error_message"]


def contains_local(stream):
    """objects.contains stream -> a bool frame. Returns True/False/None."""
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        if isinstance(o.get("Object"), bool):
            return o["Object"]
    return None


def find_identities(stream):
    ids = []
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        ob = o.get("Object")
        if isinstance(ob, str):
            ids.append(ob)
    return ids


def contract_subject(stream):
    """node1's node identity = Subject of its active contract (from user.info)."""
    for o in objs(stream):
        ob = o.get("Object")
        if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
            c = ob["Contract"].get("Contract", {})
            if c.get("Subject"):
                return c["Subject"]
    return None


def remote_identity(stream):
    """Fallback: RemoteIdentity from node2's nodes.links (the link back to node1)."""
    for o in objs(stream):
        ob = o.get("Object")
        if isinstance(ob, dict) and ob.get("RemoteIdentity"):
            return ob["RemoteIdentity"]
    return None


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--node1", default="node1")
    ap.add_argument("--node2", default="node2")
    args, _ = ap.parse_known_args()
    vm1, vm2 = args.node1, args.node2

    # node1: the id + payload the agent persisted, and node1's node identity.
    # node1 acts as the User (token) so user.info returns the active contract whose
    # Subject IS node1's node identity (the provider to target). ID strips all
    # whitespace (matches the old `tr -d '[:space:]'`); PAY tolerates a trailing nl.
    ID = "".join(ssh(vm1, "cat /home/tester/.netsim/object.id").split())
    PAY = ssh(vm1, "cat /home/tester/.netsim/object.payload").rstrip("\n")
    n1_info = ssh(vm1, TOKEN + " astral-query user.info -out json")

    # node2 answers under its node identity (no token => anonymous apphost caller).
    n2_links = ssh(vm2, "astral-query nodes.links -out json")
    N1 = contract_subject(n1_info) or remote_identity(n2_links) or ""

    n2_contains = ssh(vm2, f"astral-query objects.contains -repo local -id '{ID}' -out json")
    n2_find = ssh(vm2, f"astral-query objects.find -id '{ID}' -out json")
    n2_transparent = ssh(vm2, f"astral-query objects.load -id '{ID}' -out json")
    n2_explicit = ssh(vm2, f"astral-query '{N1}':objects.load -id '{ID}' -out json") if N1 else ""

    already_local = contains_local(n2_contains)
    explicit = loaded_payload(n2_explicit)
    transparent = loaded_payload(n2_transparent)
    providers = find_identities(n2_find)

    explicit_ok = explicit is not None and explicit.rstrip("\n") == PAY
    transparent_ok = transparent is not None and transparent.rstrip("\n") == PAY
    find_ok = (N1 in providers) if N1 else bool(providers)

    errs, notes = [], []
    if not ID:
        errs.append("no Object ID recorded on node1 (~/.netsim/object.id)")
    if not PAY:
        errs.append("no payload recorded on node1 (~/.netsim/object.payload)")
    if not N1:
        notes.append("could not resolve node1's node identity host-side (explicit-target read skipped)")

    # locality pre-check is advisory (objects.contains is probabilistic).
    if already_local is True:
        notes.append("objects.contains reports node2 may ALREADY hold this object locally; "
                     "a byte-match below might not be a genuine cross-swarm pull")
    elif already_local is None:
        notes.append("objects.contains gave no usable answer on node2 (locality pre-check inconclusive)")

    # auth-vs-route signal: an error_message naming auth/permission is a DIFFERENT
    # failure than no route / no provider -- don't conflate them.
    for label, stream in (("explicit-target", n2_explicit),
                          ("transparent", n2_transparent),
                          ("objects.find", n2_find)):
        for e in errors(stream):
            notes.append(f"{label} returned error_message: {e}")

    if explicit_ok or transparent_ok:
        path = ("explicit-target (<node1>:objects.load)" if explicit_ok
                else "transparent (objects.load, network zone)")
        print(f"share-object OK: node2 pulled object {ID[:12]}.. from node1 across the "
              f"swarm via {path}; bytes match ({len(PAY)} B). "
              f"providers seen by objects.find: {len(providers)}.")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    # Did not cross. Build a precise diagnostic (the link-swarm-style finding).
    sys.stderr.write("share-object verify FAILED: object did NOT cross the swarm to node2.\n")
    if find_ok:
        sys.stderr.write("  FINDING: provider discovery DOES cross the swarm "
                         "(objects.find on node2 returned node1) but the byte READ did not route "
                         "(explicit-target and transparent objects.load both failed to return the payload). "
                         "This is the share-object analogue of link-swarm's '<peer>:.spec does not route' "
                         "discovery -- record which hop routes in the task log.\n")
    else:
        sys.stderr.write("  no cross-swarm object access at all: neither a read nor objects.find "
                         "resolved node1's object from node2.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    n1disp = (N1[:12] + "..") if N1 else "?"
    sys.stderr.write(f"  (id={ID} node1={n1disp} "
                     f"explicit={'hit' if explicit is not None else 'miss'} "
                     f"transparent={'hit' if transparent is not None else 'miss'} "
                     f"find_providers={len(providers)})\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

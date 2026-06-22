#!/usr/bin/env python3
"""verify read-remote-object: node2 reads node1's object OVER ASTRAL.

node1 stored an object in its local repo (object-store). This confirms a peer
(node2) can obtain those exact bytes across the swarm. Host-driven, since node2 has
no operator. PRE-#348 this direction (peer reads node1) failed (route_not_found /
0 providers); this task re-probes it on current master. op_load is ungated, so a
successful route returns the bytes.

Ladder on node2 (strongest -> weakest), PASS iff node2 gets the exact stored bytes
via hop 1 or 2:
  1. explicit target  <node1-id>:objects.load -id <ID>   (query-target routing)
  2. transparent      objects.load -id <ID>              (zone-based)
  3. provider find    objects.find -id <ID>              (discovery; identities, not bytes)
"""
import argparse
import json
import subprocess
import sys


def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout (best-effort)."""
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


def loaded_payload(stream):
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
    """Fallback: RemoteIdentity from node2's nodes.links (the link to node1)."""
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

    # The object node1 stored (object-store) and node1's node identity. node1 acts as
    # the User (token from info.json) so user.info returns the contract whose Subject
    # is node1's node identity; node2's link-back is the fallback.
    info1 = info(vm1)
    ID = "".join(str(info1.get("object_id", "")).split())
    PAY = str(info1.get("object_payload", "")).rstrip("\n")
    token = info1.get("user_token", "")
    n1_info = ssh(vm1, f"export ASTRALD_APPHOST_TOKEN={token}; astral-query user.info -out json")
    n2_links = ssh(vm2, "astral-query nodes.links -out json")
    N1 = contract_subject(n1_info) or remote_identity(n2_links) or ""

    # node2 answers under its node identity (anonymous host-side caller, no token).
    n2_contains = ssh(vm2, f"astral-query objects.contains -repo local -id '{ID}' -out json")
    n2_explicit = ssh(vm2, f"astral-query '{N1}':objects.load -id '{ID}' -out json") if N1 else ""
    n2_transparent = ssh(vm2, f"astral-query objects.load -id '{ID}' -out json")
    n2_find = ssh(vm2, f"astral-query objects.find -id '{ID}' -out json")

    already_local = contains_local(n2_contains)
    explicit = loaded_payload(n2_explicit)
    transparent = loaded_payload(n2_transparent)
    providers = find_identities(n2_find)

    explicit_ok = explicit is not None and explicit.rstrip("\n") == PAY
    transparent_ok = transparent is not None and transparent.rstrip("\n") == PAY

    errs, notes = [], []
    if not ID:
        errs.append("no object_id in node1's info.json (run object-store first)")
    if not PAY:
        errs.append("no object_payload in node1's info.json")
    if not N1:
        notes.append("could not resolve node1's node identity host-side (explicit-target read skipped)")
    if already_local is True:
        notes.append("objects.contains reports node2 may ALREADY hold this object locally; "
                     "the byte-match might not be a genuine remote pull")

    if not errs and (explicit_ok or transparent_ok):
        path = ("explicit-target (<node1>:objects.load)" if explicit_ok
                else "transparent (objects.load)")
        print(f"read-remote-object OK: node2 read node1's object {ID[:12]}.. across the swarm "
              f"via {path}; bytes match ({len(PAY)} B). providers via objects.find: {len(providers)}.")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    sys.stderr.write("read-remote-object verify FAILED: node2 did NOT obtain node1's object across the swarm.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    if (N1 in providers) if N1 else bool(providers):
        sys.stderr.write("  FINDING: objects.find DID return a provider (discovery crosses) but the byte "
                         "READ did not route -- record which hop in the task log.\n")
    for label, stream in (("explicit-target", n2_explicit),
                          ("transparent", n2_transparent),
                          ("objects.find", n2_find)):
        for e in errors(stream):
            sys.stderr.write(f"  {label} error_message: {e}\n")
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

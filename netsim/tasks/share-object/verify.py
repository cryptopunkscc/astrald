#!/usr/bin/env python3
"""verify share-object: prove node2 physically holds the object node1 stored on it.

Independent host-side check (does not trust run.sh or the agent's read-back): a
repo-pinned, ungated objects.load -repo local on node2 must return the exact stored
bytes. Reaches the VMs via netsim ssh. See README.md for the full rationale.
"""
import argparse
import json
import subprocess
import sys

def ssh(vm, remote):
    """Run `netsim ssh <vm> -- <remote>` on the host; return stdout (best-effort).

    astral-query writes error_message frames to stdout (which we parse) and other
    failures (route_not_found, etc.) to stderr (which we drop).
    """
    p = subprocess.run(["netsim", "ssh", vm, "--", remote],
                       capture_output=True, text=True)
    return p.stdout


def info(vm):
    """The agent's $HOME/info.json (/home/tester/info.json) on the VM, as a dict."""
    try:
        return json.loads(ssh(vm, "cat /home/tester/info.json") or "{}") or {}
    except json.JSONDecodeError:
        return {}


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


def contract_subject(stream):
    """node identity = Subject of the active contract (from user.info)."""
    for o in objs(stream):
        ob = o.get("Object")
        if isinstance(ob, dict) and isinstance(ob.get("Contract"), dict):
            c = ob["Contract"].get("Contract", {})
            if c.get("Subject"):
                return c["Subject"]
    return None


def remote_identity(stream):
    """Fallback: RemoteIdentity from a nodes.links stream (the link to the peer)."""
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

    # What the agent persisted on node1. ID strips all whitespace; the text fields
    # tolerate a trailing newline.
    info1 = info(vm1)
    ID = "".join(str(info1.get("object_id", "")).split())
    PAY = str(info1.get("object_payload", "")).rstrip("\n")
    READBACK = str(info1.get("object_readback", "")).rstrip("\n")
    TARGET = "".join(str(info1.get("object_target", "")).split())

    # node2's real identity, resolved host-side: Subject of node2's active contract
    # (node2 answers user.info under its node identity), with node1's link-back as a
    # fallback. Used only to cross-check the node the agent claims it targeted.
    n2_info = ssh(vm2, "astral-query user.info -out json")
    n1_links = ssh(vm1, "astral-query nodes.links -out json")
    N2 = contract_subject(n2_info) or remote_identity(n1_links) or ""

    # DECISIVE: read the object straight out of node2's "local" repo (where
    # objects.store writes). Repo-pinned + ungated, so a hit proves node2 itself
    # holds the bytes -- not a network re-fetch from node1.
    n2_load = ssh(vm2, f"astral-query objects.load -id '{ID}' -repo local -out json")
    n2_contains = ssh(vm2, f"astral-query objects.contains -repo local -id '{ID}' -out json")
    got = loaded_payload(n2_load)
    held = contains_local(n2_contains)
    bytes_ok = got is not None and got.rstrip("\n") == PAY

    # Advisory: did the object ALSO land in node1's local repo? (Not required -- the
    # agent targeted node2 explicitly; a copy on node1 is fine, just noted.)
    n1_contains = ssh(vm1, f"astral-query objects.contains -repo local -id '{ID}' -out json")
    on_node1 = contains_local(n1_contains)

    errs, notes = [], []
    if not ID:
        errs.append("no object_id in node1's info.json")
    if not PAY:
        errs.append("no object_payload in node1's info.json")
    if READBACK and READBACK != PAY:
        notes.append(f"agent's own read-back != stored payload ({READBACK!r} != {PAY!r})")
    if TARGET and N2 and TARGET != N2:
        notes.append(f"agent stored on {TARGET[:12]}.. but node2's identity is {N2[:12]}.. "
                     "(agent may have targeted the wrong node)")
    elif not TARGET:
        notes.append("agent recorded no object_target in info.json")
    if on_node1 is True:
        notes.append("object is ALSO present in node1's local repo (a local copy alongside the "
                     "remote store -- not required, just noted)")

    if not errs and bytes_ok:
        tgt = (N2[:12] + "..") if N2 else (TARGET[:12] + ".." if TARGET else "node2")
        print(f"share-object OK: node1 stored object {ID[:12]}.. ON sibling {tgt} and node2's "
              f"local repo returns the exact bytes ({len(PAY)} B). "
              f"contains(local)={held}.")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    # Failure -- pinpoint what broke.
    sys.stderr.write("share-object verify FAILED: node2 does NOT hold the object in its local repo.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    if held is False:
        sys.stderr.write("  node2 objects.contains -repo local = false: the store never landed on node2 "
                         "(relay rejected, or the agent stored locally on node1 instead of targeting node2). "
                         "Check node2's journal for an objects.store and node1's for a 'query rejected'.\n")
    elif got is None:
        sys.stderr.write("  node2 objects.load -repo local returned no payload (see error frames below).\n")
    elif not bytes_ok:
        sys.stderr.write(f"  node2 returned bytes that do not match: got {got!r} != stored {PAY!r}.\n")
    # surface error frames (auth vs not-found vs repo-missing) without conflating.
    for label, stream in (("node2 load", n2_load), ("node2 contains", n2_contains)):
        for e in errors(stream):
            sys.stderr.write(f"  {label} error_message: {e}\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    n2disp = (N2[:12] + "..") if N2 else "?"
    sys.stderr.write(f"  (id={ID} node2={n2disp} target={(TARGET[:12]+'..') if TARGET else '?'} "
                     f"contains={held} load={'hit' if got is not None else 'miss'} "
                     f"on_node1={on_node1})\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

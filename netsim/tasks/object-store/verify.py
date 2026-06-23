#!/usr/bin/env python3
"""verify object-store: the stored object is present in the holder's local repo.

The agent (on node1) only stored an object on a target node (--target) and recorded
its id. Reading it back and confirming the bytes is verify's job: a repo-pinned,
ungated objects.load -repo local on the HOLDER must return the exact stored bytes.
The holder is resolved from --target: localnode/node1 -> node1 (the operator vm),
node2 -> node2. The object id comes from node1's object.json; the ground-truth
payload is the fixed payload.txt that run.sh shipped to the operator's home. Reaches
the VMs via netsim ssh.
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
    """The agent's $HOME/object.json (/home/tester/object.json) on the VM, as a dict."""
    try:
        return json.loads(ssh(vm, "cat /home/tester/object.json") or "{}") or {}
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
    """From an objects.load stream, the decoded payload string, or None."""
    for o in objs(stream):
        if o.get("Type") in ("eos", "error_message"):
            continue
        ob = o.get("Object")
        if isinstance(ob, str):
            return ob
    return None


def errors(stream):
    return [o.get("Object") for o in objs(stream) if o.get("Type") == "error_message"]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", default="node1")        # the operator; records object.json here
    ap.add_argument("--node2", default="node2")     # the peer
    ap.add_argument("--target", default="localnode")  # localnode/node1 -> node1; node2 -> node2
    args, _ = ap.parse_known_args()
    holder = args.node2 if args.target == args.node2 else args.vm

    info1 = info(args.vm)
    ID = "".join(str(info1.get("object_id", "")).split())
    # Canonical input: the exact bytes the agent was handed to store (run.sh shipped
    # payload.txt to the operator's home). Ground truth — we don't trust the agent's
    # own account of what it stored.
    PAY = (ssh(args.vm, "cat /home/tester/payload.txt") or "").rstrip("\n")

    # Decisive: re-load the object from the holder's local repo (repo-pinned + ungated)
    # and confirm the bytes match payload.txt — the read-back is verify's job, not the
    # agent's (the agent only stores and records the id).
    h_load = ssh(holder, f"astral-query objects.load -id '{ID}' -repo local -out json")
    got = loaded_payload(h_load)
    local_ok = got is not None and got.rstrip("\n") == PAY

    errs, notes = [], []
    if not ID:
        errs.append("no object_id in node1's object.json")
    if not PAY:
        errs.append("payload.txt missing on the operator (run.sh must ship it)")

    if not errs and local_ok:
        print(f"object-store OK (target={args.target}): {holder}'s local repo holds object "
              f"{ID[:12]}.. with the exact bytes ({len(PAY)} B).")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    sys.stderr.write(f"object-store verify FAILED (target={args.target}): {holder}'s local repo "
                     "does NOT hold the stored object.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    if got is None:
        sys.stderr.write(f"  objects.load -repo local on {holder} returned no payload (see errors below).\n")
    elif not local_ok:
        sys.stderr.write(f"  bytes mismatch: got {got!r} != stored {PAY!r}.\n")
    for e in errors(h_load):
        sys.stderr.write(f"  load error_message: {e}\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    sys.stderr.write(f"  (id={ID} holder={holder} load={'hit' if got is not None else 'miss'})\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

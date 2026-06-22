#!/usr/bin/env python3
"""verify object-store: node1 stored an object in its local repo and can read it back.

Independent host-side check (does not trust run.sh or the agent's read-back): a
repo-pinned, ungated objects.load -repo local on node1 must return the exact stored
bytes. Reaches the VM via netsim ssh.
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
    ap.add_argument("--vm", default="node1")
    args, _ = ap.parse_known_args()
    vm = args.vm

    info1 = info(vm)
    ID = "".join(str(info1.get("object_id", "")).split())
    PAY = str(info1.get("object_payload", "")).rstrip("\n")
    READBACK = str(info1.get("object_readback", "")).rstrip("\n")

    # Decisive: re-load the object from node1's local repo (repo-pinned + ungated).
    n1_load = ssh(vm, f"astral-query objects.load -id '{ID}' -repo local -out json")
    got = loaded_payload(n1_load)
    local_ok = got is not None and got.rstrip("\n") == PAY

    errs, notes = [], []
    if not ID:
        errs.append("no object_id in node1's info.json")
    if not PAY:
        errs.append("no object_payload in node1's info.json")
    if READBACK and READBACK != PAY:
        notes.append(f"agent's own read-back != stored payload ({READBACK!r} != {PAY!r})")

    if not errs and local_ok:
        print(f"object-store OK: node1 stored object {ID[:12]}.. and its local repo "
              f"returns the exact bytes ({len(PAY)} B).")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    sys.stderr.write("object-store verify FAILED: node1 could not re-load its own stored object.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    if got is None:
        sys.stderr.write("  objects.load -repo local returned no payload (see error frames below).\n")
    elif not local_ok:
        sys.stderr.write(f"  bytes mismatch: got {got!r} != stored {PAY!r}.\n")
    for e in errors(n1_load):
        sys.stderr.write(f"  load error_message: {e}\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    sys.stderr.write(f"  (id={ID} load={'hit' if got is not None else 'miss'})\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

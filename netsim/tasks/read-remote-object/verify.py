#!/usr/bin/env python3
"""verify read-remote-object: node1 read the peer's object over astral.

object-store --target node2 put the object on the peer (node2) and recorded
object_id + object_payload in node1's info.json; read-remote-object's agent (on
node1, as the User) read it back from the peer and recorded object_remote.

Independent host-side check: re-read the peer's object AS THE USER (node1 holds the
token) via <peer>:objects.load and assert the bytes equal the stored payload — this
is the authenticated, routable direction. Also cross-checks the agent's recorded
read. Reaches the VMs via netsim ssh.
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
    ap.add_argument("--vm", default="node1")      # operator; reads as the User
    ap.add_argument("--peer", default="node2")    # the node holding the object (alias)
    args, _ = ap.parse_known_args()

    info1 = info(args.vm)
    ID = "".join(str(info1.get("object_id", "")).split())
    PAY = str(info1.get("object_payload", "")).rstrip("\n")
    REMOTE = str(info1.get("object_remote", ""))
    token = info1.get("user_token", "")

    # Independent: node1, as the User, reads the peer's object over astral. This is
    # authenticated (token), so the query keeps the network zone and routes to the peer.
    tok = f"export ASTRALD_APPHOST_TOKEN={token};" if token else ""
    out = ssh(args.vm, f"{tok} astral-query {args.peer}:objects.load -id '{ID}' -out json")
    got = loaded_payload(out)
    read_ok = got is not None and got.rstrip("\n") == PAY

    errs, notes = [], []
    if not ID:
        errs.append("no object_id in node1's info.json (object-store --target node2 must run first)")
    if not PAY:
        errs.append("no object_payload in node1's info.json")
    if not token:
        errs.append("no user_token in node1's info.json (can't read the peer as the User)")
    if not REMOTE:
        notes.append("agent recorded no object_remote (the agent's own read)")
    elif PAY and PAY not in REMOTE:
        notes.append(f"agent's recorded read does not contain the payload ({REMOTE!r})")

    if not errs and read_ok:
        print(f"read-remote-object OK: node1 (as User) read object {ID[:12]}.. from "
              f"{args.peer} over astral; bytes match ({len(PAY)} B).")
        for n in notes:
            sys.stderr.write(f"  note: {n}\n")
        return 0

    sys.stderr.write(f"read-remote-object verify FAILED: node1 could not read the object from "
                     f"{args.peer} over astral.\n")
    for e in errs:
        sys.stderr.write(f"  - {e}\n")
    if got is None:
        sys.stderr.write(f"  {args.peer}:objects.load (as User) returned no payload "
                         "(route_not_found means the read didn't route — check auth/zone).\n")
    elif not read_ok:
        sys.stderr.write(f"  bytes mismatch: got {got!r} != stored {PAY!r}.\n")
    for e in errors(out):
        sys.stderr.write(f"  load error_message: {e}\n")
    for n in notes:
        sys.stderr.write(f"  note: {n}\n")
    sys.stderr.write(f"  (id={ID} peer={args.peer} read={'hit' if got is not None else 'miss'})\n")
    return 1


if __name__ == "__main__":
    sys.exit(main())

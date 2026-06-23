#!/usr/bin/env python3
"""verify enable-tor: each target VM runs Tor and saved its own onion endpoint.

Independent host-side check (does not trust run.sh): on each VM the tor service is active,
/root/tor.json holds an onion endpoint, and that saved onion matches what astrald actually
advertises now (nodes.resolve_endpoints -id localnode). Reaches the VMs via netsim ssh.
"""
import argparse
import json
import subprocess
import sys


def ssh(vm, remote):
    p = subprocess.run(["netsim", "ssh", vm, "--", remote], capture_output=True, text=True)
    return p.stdout


def all_running_vms():
    out = subprocess.run(["netsim", "vm", "ls", "--json"], capture_output=True, text=True).stdout
    try:
        return [v["hostname"] for v in json.loads(out or "[]") if v.get("state") == "running"]
    except json.JSONDecodeError:
        return []


def endpoint_addr(ep):
    if isinstance(ep, str):
        return ep
    if isinstance(ep, dict):
        o = ep.get("Object")
        return o if isinstance(o, str) else ""
    return ""


def live_onion(vm):
    """The onion address astrald advertises now (resolve_endpoints -id localnode), or None."""
    stream = ssh(vm, "astral-query nodes.resolve_endpoints -id localnode -out json")
    for ln in (stream or "").splitlines():
        ln = ln.strip()
        if not ln:
            continue
        try:
            o = json.loads(ln)
        except json.JSONDecodeError:
            continue
        a = endpoint_addr((o.get("Object") or {}).get("Endpoint"))
        if ".onion" in a:
            return a
    return None


def saved(vm):
    """The contents of /root/tor.json on the VM, as a dict."""
    try:
        return json.loads(ssh(vm, "cat /root/tor.json") or "{}") or {}
    except json.JSONDecodeError:
        return {}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--vm", action="append", default=[])
    args, _ = ap.parse_known_args()
    vms = args.vm or all_running_vms()
    if not vms:
        sys.stderr.write("enable-tor verify FAILED: no VMs to verify\n")
        return 1

    bad = False
    for vm in vms:
        tor_active = ssh(vm, "systemctl is-active tor 2>/dev/null").strip() == "active"
        file_onion = str(saved(vm).get("onion", ""))
        live = live_onion(vm)

        errs = []
        if not tor_active:
            errs.append("the tor service is not active")
        if not file_onion:
            errs.append("no onion in /root/tor.json")
        if not live:
            errs.append("astrald advertises no onion (resolve_endpoints -id localnode)")
        if file_onion and live and file_onion != live:
            errs.append(f"saved onion {file_onion} != live onion {live}")

        if errs:
            bad = True
            sys.stderr.write(f"enable-tor verify FAILED on {vm}:\n")
            for e in errs:
                sys.stderr.write(f"  - {e}\n")
        else:
            print(f"enable-tor OK: {vm} runs tor and saved its onion {file_onion} to /root/tor.json")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())

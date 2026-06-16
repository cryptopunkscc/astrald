#!/bin/sh
# verify install-astrald (same args as run.sh): on every target VM the astrald
# service must be active AND the node must answer its local API. This is an
# INDEPENDENT re-check -- it re-derives the VM list and re-probes the node; it
# does not trust run.sh's output.
set -eu
VMS=""
while [ $# -gt 0 ]; do
  case "$1" in
    --vm)  VMS="${VMS:+$VMS }$2"; shift 2 ;;
    --ref) shift 2 ;;
    *) shift ;;
  esac
done
if [ -z "$VMS" ]; then
  VMS=$(netsim vm ls --json | python3 -c \
    'import json,sys; print(" ".join(v["hostname"] for v in json.load(sys.stdin) if v["state"]=="running"))')
fi
[ -n "$VMS" ] || { echo "no running VMs to verify" >&2; exit 1; }

# $VMS is a space-separated list -> intentional word-splitting
# shellcheck disable=SC2086
for vm in $VMS; do
  # single-quoted: $(hostname) must expand on the guest, not the host
  # shellcheck disable=SC2016
  netsim ssh "$vm" -- 'systemctl is-active --quiet astrald \
      && timeout 5 astral-query localnode:.spec -out json >/dev/null 2>&1 \
      && echo "$(hostname): astrald healthy"' \
    || { echo "astrald NOT healthy on $vm" >&2; exit 1; }
done
echo "verified astrald on: $VMS"

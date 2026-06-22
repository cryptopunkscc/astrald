#!/bin/sh
# object-store: have node1 store an astral object in its OWN local repo and read it
# back, driven by the Qwen Code agent running INSIDE node1.
#   object-store [--vm <host>]      (default: node1 — the VM carrying Qwen)
#
# Runs ON THE HOST (cwd = simulation root). Same mechanic as bootstrap-user-*: tiny
# script, thin prompt, intelligence in the agent's astral-agent skill. The agent
# exercises the basic LOCAL object operations — store a payload, surface its Object
# ID, load it back by that id, confirm the round-trip. verify.py then INDEPENDENTLY
# re-reads the object from node1's local repo. The whole remote program travels as
# ONE argv to `netsim ssh`; the prompt rides along base64-encoded.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: object-store [--vm <host>]" >&2; exit 64 ;;
  esac
done

# CDPATH= is an intentional one-shot env prefix for cd, not an assignment
# shellcheck disable=SC1007
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
[ -f "$here/prompt.md" ] || { echo "missing $here/prompt.md" >&2; exit 1; }
prompt_b64=$(base64 -w0 "$here/prompt.md")   # GNU coreutils; -w0 = single line

REMOTE_BODY=$(cat <<'EOS'
set -eu
d=/home/tester/.netsim
mkdir -p "$d"
printf '%s' "$prompt_b64" | base64 -d > "$d/object-store.prompt"
chown -R tester:tester "$d"

su - tester -c 'qwen -y "$(cat /home/tester/.netsim/object-store.prompt)"' \
   > "$d/object-store.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/object-store.log" >&2
     exit 1
   }

# Cheap smoke-check; verify.py does the authoritative, independent check. The agent
# records its outputs in $HOME/info.json (/home/tester/info.json).
oid=$(python3 -c 'import json;print(json.load(open("/home/tester/info.json")).get("object_id",""))' 2>/dev/null || true)
opay=$(python3 -c 'import json;print(json.load(open("/home/tester/info.json")).get("object_payload",""))' 2>/dev/null || true)
orb=$(python3 -c 'import json;print(json.load(open("/home/tester/info.json")).get("object_readback",""))' 2>/dev/null || true)
[ -n "$oid" ]  || { echo "agent recorded no object_id in /home/tester/info.json on $(hostname)" >&2; exit 1; }
[ -n "$opay" ] || { echo "agent recorded no object_payload on $(hostname)" >&2; exit 1; }
[ -n "$orb" ]  || { echo "agent recorded no object_readback on $(hostname)" >&2; exit 1; }
case "$oid" in
  data1*) : ;;
  *) echo "WARNING $(hostname): object_id does not look like a data1… Object ID (verify.py decides)" >&2 ;;
esac
[ "$opay" = "$orb" ] || echo "WARNING $(hostname): agent read-back != stored payload (verify.py decides)" >&2
echo "object-store: agent finished on $(hostname); stored+read object $oid"
EOS
)

echo "object-store: driving Qwen operator on $VM ..."
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "object-store: done on $VM"

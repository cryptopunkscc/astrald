#!/bin/sh
# share-object: have the operator node store an astral object, driven by the Qwen
# Code agent running INSIDE node1 (already a User node in one swarm with node2 —
# default starting stage: astrald-swarm).
#   share-object [--vm <host>]      (default: node1 — the VM carrying Qwen)
#
# Runs ON THE HOST (cwd = simulation root). Same mechanic as bootstrap-user /
# link-swarm: tiny script, thin prompt, intelligence in the agent's astral-agent
# skill. The agent does ONLY the store half (store a payload, surface its Object
# ID); the cross-swarm fetch from node2 is left entirely to verify.sh, which can
# address the sibling deterministically. The whole remote program travels as ONE
# argv to `netsim ssh`; the prompt rides along base64-encoded so a multi-line
# file never fights shell quoting.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: share-object [--vm <host>]" >&2; exit 64 ;;
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
printf '%s' "$prompt_b64" | base64 -d > "$d/share-object.prompt"
chown -R tester:tester "$d"

# Run the agent as `tester` (qwen is installed for that user), non-interactively.
# Invocation matches what was validated for bootstrap-user / link-swarm: one-shot
# positional prompt + `-y` (auto-approve).
su - tester -c 'qwen -y "$(cat /home/tester/.netsim/share-object.prompt)"' \
   > "$d/share-object.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/share-object.log" >&2
     exit 1
   }

# Cheap smoke-check; verify.sh does the authoritative, independent cross-swarm
# check. The agent must have recorded an Object ID and the payload it stored.
[ -s "$d/object.id" ]      || { echo "agent recorded no Object ID on $(hostname) (~/.netsim/object.id)" >&2; exit 1; }
[ -s "$d/object.payload" ] || { echo "agent recorded no payload on $(hostname) (~/.netsim/object.payload)" >&2; exit 1; }
case "$(cat "$d/object.id")" in
  data1*) : ;;
  *) echo "WARNING $(hostname): object.id does not look like a data1… Object ID (verify.sh decides)" >&2 ;;
esac
echo "share-object: agent finished on $(hostname); stored object $(cat "$d/object.id")"
EOS
)

echo "share-object: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "share-object: done on $VM"

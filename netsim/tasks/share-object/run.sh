#!/bin/sh
# share-object: have node1 store an astral object ON its swarm sibling (node2) and
# read it back, driven by the Qwen Code agent running INSIDE node1 (already a User
# node in one swarm with node2 — default starting stage: astrald-swarm).
#   share-object [--vm <host>]      (default: node1 — the VM carrying Qwen)
#
# Runs ON THE HOST (cwd = simulation root). Same mechanic as bootstrap-user /
# link-swarm: tiny script, thin prompt, intelligence in the agent's astral-agent
# skill. The agent stores a payload ON THE OTHER node — addressing the sibling
# explicitly as the query target (<node2>:objects.store) — then loads it back from
# that node and confirms the bytes round-trip. verify.py then INDEPENDENTLY
# confirms node2 physically holds the object in its LOCAL repo with matching bytes
# (objects.contains/load -repo local on node2). The whole remote program travels as
# ONE argv to `netsim ssh`; the prompt rides along base64-encoded so a multi-line
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

# Cheap smoke-check; verify.py does the authoritative, independent check (node2
# physically holds the object in its local repo). The agent must have recorded an
# Object ID, the payload it stored, the bytes it read back, and the node it stored
# it on.
[ -s "$d/object.id" ]       || { echo "agent recorded no Object ID on $(hostname) (~/.netsim/object.id)" >&2; exit 1; }
[ -s "$d/object.payload" ]  || { echo "agent recorded no payload on $(hostname) (~/.netsim/object.payload)" >&2; exit 1; }
[ -s "$d/object.readback" ] || { echo "agent recorded no read-back on $(hostname) (~/.netsim/object.readback)" >&2; exit 1; }
[ -s "$d/object.target" ]   || { echo "agent recorded no target node on $(hostname) (~/.netsim/object.target)" >&2; exit 1; }
case "$(cat "$d/object.id")" in
  data1*) : ;;
  *) echo "WARNING $(hostname): object.id does not look like a data1… Object ID (verify.py decides)" >&2 ;;
esac
# Advisory: the agent's own round-trip should already match (verify.py re-checks).
[ "$(cat "$d/object.payload")" = "$(cat "$d/object.readback")" ] \
  || echo "WARNING $(hostname): agent read-back != stored payload (verify.py decides)" >&2
echo "share-object: agent finished on $(hostname); stored object $(cat "$d/object.id") on $(cat "$d/object.target")"
EOS
)

echo "share-object: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "share-object: done on $VM"

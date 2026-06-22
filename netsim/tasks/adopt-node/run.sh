#!/bin/sh
# adopt-node: adopt the second node into the User's swarm, driven by the Qwen
# Code agent running INSIDE node1 (which is already a User node from
# bootstrap-user-software-key — default starting stage: astrald-user).
#   adopt-node [--vm <host>]      (default: node1 — the VM carrying Qwen)
#
# Runs ON THE HOST (cwd = simulation root). Same mechanic as bootstrap-user-software-key:
# tiny script, thin prompt, intelligence in the agent's astral-agent skill. The
# whole remote program travels as ONE argv to `netsim ssh`; the prompt rides
# along base64-encoded so a multi-line file never fights shell quoting.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: adopt-node [--vm <host>]" >&2; exit 64 ;;
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
printf '%s' "$prompt_b64" | base64 -d > "$d/adopt-node.prompt"
chown -R tester:tester "$d"

# Run the agent as `tester` (qwen is installed for that user), non-interactively.
# Invocation matches what was validated for bootstrap-user-software-key: one-shot positional
# prompt + `-y` (auto-approve).
su - tester -c 'qwen -y "$(cat /home/tester/.netsim/adopt-node.prompt)"' \
   > "$d/adopt-node.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/adopt-node.log" >&2
     exit 1
   }

# Soft smoke-check only (verify.sh is the authoritative, independent check).
# node1 already holds a User token from bootstrap-user-software-key, so we can peek at the
# swarm here; don't fail the run on a shape mismatch — leave the verdict to
# verify.sh.  CONFIRM the user.swarm_status JSON field for a linked sibling.
tok="$d/user.token"
if [ -s "$tok" ]; then
  if ASTRALD_APPHOST_TOKEN=$(cat "$tok") astral-query user.swarm_status -out json 2>/dev/null \
       | grep -q '"Linked":true'; then
    echo "adopt-node: $(hostname) reports a linked sibling"
  else
    echo "adopt-node: WARNING $(hostname) shows no linked sibling yet (verify.sh decides)" >&2
  fi
fi
echo "adopt-node: agent finished on $(hostname)"
EOS
)

echo "adopt-node: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "adopt-node: done on $VM"

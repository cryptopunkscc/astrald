#!/bin/sh
# bootstrap-user: turn the operator node into a User-controlled node, driven by
# the Qwen Code agent running INSIDE the VM.
#   bootstrap-user [--vm <host>]      (default: node1 — the VM carrying Qwen)
#
# Runs ON THE HOST (cwd = simulation root). This script is deliberately tiny: it
# ships prompt.md to the agent on the guest and lets the agent do the astral
# work via astral-query against the local node API. The intelligence lives in
# the prompt and — by design — in the agent's astral-agent skill, not here. The
# whole remote program travels as ONE argv to `netsim ssh` (no reliance on stdin
# forwarding); the prompt rides along base64-encoded so a multi-line file never
# fights shell quoting.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: bootstrap-user [--vm <host>]" >&2; exit 64 ;;
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
printf '%s' "$prompt_b64" | base64 -d > "$d/bootstrap-user.prompt"
chown -R tester:tester "$d"

# Run the agent as `tester` (qwen is installed for that user), non-interactively.
# Invocation matches what was validated against the live lab: one-shot positional
# prompt + `-y` (auto-approve). The prompt is passed positionally via command
# substitution; the substituted text is used literally (not re-scanned), so the
# backticks and $-signs inside it are safe.
su - tester -c 'qwen -y "$(cat /home/tester/.netsim/bootstrap-user.prompt)"' \
   > "$d/bootstrap-user.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/bootstrap-user.log" >&2
     exit 1
   }

# Cheap smoke-check; verify.sh does the authoritative, independent check.
[ -s "$d/user.id" ] || { echo "agent recorded no User id on $(hostname)" >&2; exit 1; }
echo "bootstrap-user: agent finished on $(hostname); User id $(cat "$d/user.id")"
EOS
)

echo "bootstrap-user: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "bootstrap-user: done on $VM"

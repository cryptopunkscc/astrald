#!/bin/sh
# import-user: configure the operator node as a User node from an EXISTING software
# User key — the User's BIP-39 mnemonic (env ASTRAL_USER_MNEMONIC) is derived
# instead of minting fresh entropy. Driven by the Qwen Code agent in the VM.
#   import-user [--vm <host>]      (default: node1 — the VM carrying Qwen)
#   env: ASTRAL_USER_MNEMONIC (required)   ASTRAL_USER_ID (optional; verify.sh asserts it)
#
# Drop-in alternative to bootstrap-user. Runs ON THE HOST (cwd = simulation root):
# substitutes the mnemonic into prompt.md, base64-ships the prompt to the agent over
# one `netsim ssh` argv, and runs `qwen -y`. Intelligence lives in the prompt and
# the agent's astral-agent skill, not here.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: import-user [--vm <host>]   (env ASTRAL_USER_MNEMONIC required)" >&2; exit 64 ;;
  esac
done

[ -n "${ASTRAL_USER_MNEMONIC:-}" ] \
  || { echo "set ASTRAL_USER_MNEMONIC to the User's BIP-39 mnemonic seed phrase" >&2; exit 64; }

# CDPATH= is an intentional one-shot env prefix for cd, not an assignment
# shellcheck disable=SC1007
here=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
[ -f "$here/prompt.md" ] || { echo "missing $here/prompt.md" >&2; exit 1; }
# Substitute the mnemonic into the prompt template (BIP-39 words are [a-z ] only,
# so no sed-delimiter or regex-metachar hazard).
prompt=$(sed "s|__MNEMONIC__|$ASTRAL_USER_MNEMONIC|" "$here/prompt.md")
prompt_b64=$(printf '%s' "$prompt" | base64 -w0)   # GNU coreutils; -w0 = single line

REMOTE_BODY=$(cat <<'EOS'
set -eu
d=/home/tester/.netsim
mkdir -p "$d"
printf '%s' "$prompt_b64" | base64 -d > "$d/import-user.prompt"
chown -R tester:tester "$d"

# Run the agent as `tester` (qwen is installed for that user), non-interactively:
# one-shot positional prompt + `-y` (auto-approve).
su - tester -c 'qwen -y "$(cat /home/tester/.netsim/import-user.prompt)"' \
   > "$d/import-user.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/import-user.log" >&2
     exit 1
   }

# Cheap smoke-check; verify.sh does the authoritative, independent check. The agent
# records its outputs in $HOME/info.json (/home/tester/info.json).
uid=$(python3 -c 'import json;print(json.load(open("/home/tester/info.json")).get("user_id",""))' 2>/dev/null || true)
[ -n "$uid" ] || { echo "agent recorded no user_id in /home/tester/info.json on $(hostname)" >&2; exit 1; }
echo "import-user: agent finished on $(hostname); User id $uid"
EOS
)

echo "import-user: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "import-user: done on $VM"

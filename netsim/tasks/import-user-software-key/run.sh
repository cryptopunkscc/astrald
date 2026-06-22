#!/bin/sh
# import-user-software-key: configure the operator node as a User node from an
# EXISTING software User key — the User's BIP-39 mnemonic is embedded in prompt.md
# and derived instead of minting fresh entropy. Driven by the Qwen Code agent.
#   import-user-software-key [--vm <host>]   (default: node1 — the VM carrying Qwen)
#   env: ASTRAL_USER_ID (optional; verify.sh asserts the derived id matches it)
#
# Drop-in alternative to bootstrap-user-software-key. Runs ON THE HOST (cwd =
# simulation root): base64-ships prompt.md to the agent over one `netsim ssh` argv
# and runs `qwen -y`. Intelligence lives in the prompt and the agent's astral-agent
# skill, not here.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    *)    echo "usage: import-user-software-key [--vm <host>]" >&2; exit 64 ;;
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
printf '%s' "$prompt_b64" | base64 -d > "$d/import-user-software-key.prompt"
chown -R tester:tester "$d"

# Run the agent as `tester` (qwen is installed for that user), non-interactively:
# one-shot positional prompt + `-y` (auto-approve).
su - tester -c 'qwen -y "$(cat /home/tester/.netsim/import-user-software-key.prompt)"' \
   > "$d/import-user-software-key.log" 2>&1 || {
     echo "qwen run failed on $(hostname); tail of log:" >&2
     tail -n 40 "$d/import-user-software-key.log" >&2
     exit 1
   }

# Cheap smoke-check; verify.sh does the authoritative, independent check. The agent
# records its outputs in $HOME/user.json (/home/tester/user.json).
uid=$(python3 -c 'import json;print(json.load(open("/home/tester/user.json")).get("user_id",""))' 2>/dev/null || true)
[ -n "$uid" ] || { echo "agent recorded no user_id in /home/tester/user.json on $(hostname)" >&2; exit 1; }
echo "import-user-software-key: agent finished on $(hostname); User id $uid"
EOS
)

echo "import-user-software-key: driving Qwen operator on $VM ..."
# assignment prefix carries the prompt to the guest; body re-parses it
# shellcheck disable=SC2029
netsim ssh "$VM" -- "prompt_b64='$prompt_b64'; $REMOTE_BODY"
echo "import-user-software-key: done on $VM"

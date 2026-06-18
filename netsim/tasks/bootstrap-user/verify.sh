#!/bin/sh
# verify bootstrap-user (same args as run.sh): the target node must be a
# User-controlled node. INDEPENDENT re-check -- it does not trust run.sh's
# output: it reads the persisted User credentials, acts AS the User, and asserts
# the node answers as a user node. user.info itself rejects (code 2) when there
# is no active contract, so a successful call IS the proof.
set -eu

VM="node1"
while [ $# -gt 0 ]; do
  case "$1" in
    --vm) VM=$2; shift 2 ;;
    *)    shift ;;
  esac
done

REMOTE_CHECK=$(cat <<'EOS'
set -eu
d=/home/tester/.netsim
[ -s "$d/user.id" ]    || { echo "no recorded User id on $(hostname)"    >&2; exit 1; }
[ -s "$d/user.token" ] || { echo "no recorded User token on $(hostname)" >&2; exit 1; }
uid=$(cat "$d/user.id")
ASTRALD_APPHOST_TOKEN=$(cat "$d/user.token"); export ASTRALD_APPHOST_TOKEN

# acting as the User: whoami must report the User identity
who=$(astral-query apphost.whoami -out json) \
  || { echo "apphost.whoami failed on $(hostname)" >&2; exit 1; }
echo "$who" | grep -q "$uid" \
  || { echo "whoami != User id on $(hostname): $who" >&2; exit 1; }

# active contract present (user.info rejects with code 2 if none)
astral-query user.info -out json \
  || { echo "user.info failed on $(hostname) -- no active contract?" >&2; exit 1; }

echo "$(hostname): user node OK (User $uid)"
EOS
)

netsim ssh "$VM" -- "$REMOTE_CHECK" \
  || { echo "bootstrap-user verify FAILED on $VM" >&2; exit 1; }
echo "verified user node on: $VM"

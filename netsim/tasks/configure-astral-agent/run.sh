#!/bin/sh
# configure-astral-agent: install the astral-agent skill into the Qwen Code
# operator by having the VM clone the (private) satforgedev/skills repo with an
# injected deploy key and run the linker itself.
#   configure-astral-agent [--vm <host>] [--user <name>]
# Default: --vm node1 --user tester (the operator created by install-qwen-code).
#
# The HOST owns the deploy key; the VM never needs GitHub credentials of its own.
# run.sh reads the private key path from $SATFORGE_SKILLS_DEPLOY_KEY, base64-ships
# it in over a single `netsim ssh` argv, and the guest then:
#   1. installs the key for the operator and clones
#      git@github.com:satforgedev/skills (parent over SSH via the deploy key;
#      the astral-docs submodule is public HTTPS, so it needs no key),
#   2. builds the satforge-skills linker (Go is already on the node from
#      install-astrald),
#   3. runs `link astral-agent --target qwen` -> ~<user>/.qwen/skills/astral-agent.
#
# NOTE: for now the deploy key is LEFT in the VM (simpler; lets the operator
# re-clone/pull skills later), which means it also lives in the saved snapshot.
# We may switch to wiping the key before the snapshot if that exposure matters.
set -eu

VM=node1
USER_NAME=tester
while [ $# -gt 0 ]; do
  case "$1" in
    --vm)   [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VM=$2; shift 2 ;;
    --user) [ $# -ge 2 ] || { echo "need name after --user" >&2; exit 64; }; USER_NAME=$2; shift 2 ;;
    *) echo "usage: configure-astral-agent [--vm <host>] [--user <name>]" >&2; exit 64 ;;
  esac
done

REPO=${SATFORGE_SKILLS_REPO:-git@github.com:satforgedev/skills}
KEY=${SATFORGE_SKILLS_DEPLOY_KEY:-}
[ -n "$KEY" ] || { echo "set SATFORGE_SKILLS_DEPLOY_KEY to the deploy key path for $REPO" >&2; exit 1; }
[ -r "$KEY" ] || { echo "deploy key not readable: $KEY" >&2; exit 1; }
key_b64=$(base64 -w0 "$KEY")

REMOTE_BODY=$(cat <<'EOS'
set -eu
home=$(getent passwd "$u" | cut -d: -f6)
[ -n "$home" ] || { echo "user '$u' not found on $(hostname)" >&2; exit 1; }
command -v git >/dev/null 2>&1 || { echo "git missing on $(hostname)" >&2; exit 1; }

install -d -m 700 -o "$u" -g "$u" "$home/.ssh" "$home/.netsim"
printf '%s' "$key_b64" | base64 -d > "$home/.ssh/skills_deploy"
chmod 600 "$home/.ssh/skills_deploy"
chown "$u:$u" "$home/.ssh/skills_deploy"

# Guest-side provisioning, run as the operator. Quoted heredoc: fully literal;
# repo + ref arrive as positional args. github's host key is auto-accepted on
# first connect. If outbound SSH:22 is ever blocked, switch the URL to
# ssh://git@ssh.github.com:443/satforgedev/skills.
cat > "$home/.netsim/setup-skill.sh" <<'SCRIPT'
#!/bin/sh
set -eu
export PATH=/usr/local/go/bin:$PATH
export GIT_SSH_COMMAND="ssh -i $HOME/.ssh/skills_deploy -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new"
repo=$1
src=$HOME/satforge-skills
[ -d "$src/.git" ] || git clone --recurse-submodules "$repo" "$src"
cd "$src"
git pull --ff-only --quiet 2>/dev/null || true
git submodule update --init --recursive --quiet
go build -C bin/satforge-skills -o satforge-skills .
bin="$src/bin/satforge-skills/satforge-skills"
"$bin" unlink astral-agent --target qwen >/dev/null 2>&1 || true   # idempotent re-run
"$bin" link astral-agent --target qwen
SCRIPT
chown "$u:$u" "$home/.netsim/setup-skill.sh"

su - "$u" -c "sh '$home/.netsim/setup-skill.sh' '$repo'"
echo "configure-astral-agent: $(hostname) cloned skills + linked astral-agent (deploy key left in place)"
EOS
)

echo "configure-astral-agent: injecting deploy key + linking on $VM (user $USER_NAME) ..."
netsim ssh "$VM" -- "u='$USER_NAME' key_b64='$key_b64' repo='$REPO'; $REMOTE_BODY"
echo "configure-astral-agent: done on $VM"

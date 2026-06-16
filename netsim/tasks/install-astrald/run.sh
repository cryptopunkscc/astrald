#!/bin/sh
# install-astrald: build astrald from source, install it as a systemd service on VMs.
#   install-astrald [--vm <host>]... [--ref <git-ref>]
# No --vm  ->  every running VM in the simulation.
#
# Runs ON THE HOST (cwd = simulation root). Reaches each VM with a single
# `netsim ssh <host> -- <one-argv>` call: the whole remote script travels as ONE
# argument (assignment prefix + single-quoted heredoc body, so host-side $... are
# left for the guest to expand). ssh lands as root on the guest.
set -eu
REPO="https://github.com/cryptopunkscc/astrald"
GO_VERSION="1.25.1"          # must be >= 1.25.0 (astrald go.mod); pin to current 1.25.x
REF=""

VMS=""
while [ $# -gt 0 ]; do
  case "$1" in
    --vm)  [ $# -ge 2 ] || { echo "need host after --vm" >&2; exit 64; }; VMS="${VMS:+$VMS }$2"; shift 2 ;;
    --ref) [ $# -ge 2 ] || { echo "need ref after --ref"  >&2; exit 64; }; REF=$2;       shift 2 ;;
    *) echo "usage: install-astrald [--vm <host>]... [--ref <git-ref>]" >&2; exit 64 ;;
  esac
done
if [ -z "$VMS" ]; then
  VMS=$(netsim vm ls --json | python3 -c \
    'import json,sys; print(" ".join(v["hostname"] for v in json.load(sys.stdin) if v["state"]=="running"))')
fi
[ -n "$VMS" ] || { echo "no running VMs" >&2; exit 1; }

REMOTE_BODY=$(cat <<'EOS'
set -eu
export DEBIAN_FRONTEND=noninteractive

# deps: git + curl (Go comes from the official tarball, not apt -> need >= 1.25)
need=""; command -v git  >/dev/null 2>&1 || need="$need git"
         command -v curl >/dev/null 2>&1 || need="$need curl"
if [ -n "$need" ]; then
    apt-get -qq -o DPkg::Lock::Timeout=120 update
    apt-get -qq -y -o DPkg::Lock::Timeout=120 install $need ca-certificates >/dev/null
fi
if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go$go_ver "; then
    case "$(uname -m)" in
        x86_64)  ga=amd64 ;; aarch64) ga=arm64 ;;
        *) echo "unsupported arch $(uname -m)" >&2; exit 1 ;;
    esac
    t=$(mktemp); curl -fsSL -o "$t" "https://go.dev/dl/go${go_ver}.linux-${ga}.tar.gz"
    rm -rf /usr/local/go; tar -C /usr/local -xzf "$t"; rm -f "$t"
fi
export PATH=/usr/local/go/bin:$PATH CGO_ENABLED=0

# build (plain clone, NO --recursive; subpackages need the ./ prefix)
src=/opt/astrald-src
[ -d "$src/.git" ] || git clone --depth 1 ${ref:+--branch "$ref"} "$repo" "$src"
cd "$src"
go build -o /usr/local/bin/astrald      ./cmd/astrald
go build -o /usr/local/bin/astral-query ./cmd/astral-query

# run as a service: explicit -root and HOME (default paths break without HOME)
install -d -m 700 /var/lib/astrald
cat > /etc/systemd/system/astrald.service <<UNIT
[Unit]
Description=astral daemon
[Service]
ExecStart=/usr/local/bin/astrald -root /var/lib/astrald
Environment=HOME=/root
Restart=on-failure
KillSignal=SIGINT
[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload
systemctl enable --now astrald

# confirm it built AND runs: wait for the apphost listener, then probe the API
ok=
for _ in 1 2 3 4 5 6 7 8 9 10; do
    if systemctl is-active --quiet astrald && timeout 5 astral-query localnode:.spec -out json >/dev/null 2>&1; then
        ok=1; break
    fi
    sleep 1
done
[ -n "$ok" ] || { echo "astrald did not come up on $(hostname)" >&2; exit 1; }

# stop it so netsim snapshots an idle guest; the unit stays enabled and
# autostarts when the stage boots. a running daemon keeps dirtying RAM and can
# stall the live snapshot (the qmp timeout).
systemctl stop astrald
echo "astrald installed and verified; enabled, stopped for snapshot on $(hostname)"
EOS
)

# $VMS is a space-separated list -> intentional word-splitting
# shellcheck disable=SC2086
for vm in $VMS; do
  echo "installing astrald on $vm ..."
  netsim ssh "$vm" -- "repo='$REPO' ref='$REF' go_ver='$GO_VERSION'; $REMOTE_BODY"
done
